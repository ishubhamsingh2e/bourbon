package router

import (
	"net/http"
	"regexp"
	"strings"
	"sync"
)

type nodeType uint8

const (
	nodeStatic nodeType = iota
	nodeParam
	nodeCatchAll
)

type routeEntry struct {
	handler http.Handler
	pattern string
}

type trieNode struct {
	segment     string
	paramName   string
	regex       *regexp.Regexp
	nodeType    nodeType
	routes      map[string]routeEntry // HTTP Method -> routeEntry
	children    []*trieNode
	mountServer http.Handler
}

func newTrieNode(segment string, nType nodeType, paramName string, re *regexp.Regexp) *trieNode {
	return &trieNode{
		segment:   segment,
		nodeType:  nType,
		paramName: paramName,
		regex:     re,
		routes:    make(map[string]routeEntry),
		children:  make([]*trieNode, 0),
	}
}

// Tree represents the Radix Trie routing structure.
type Tree struct {
	mu   sync.RWMutex
	root *trieNode
}

// NewTree initializes an empty Radix Trie.
func NewTree() *Tree {
	return &Tree{
		root: newTrieNode("", nodeStatic, "", nil),
	}
}

func parseSegment(seg string) (string, nodeType, string, *regexp.Regexp) {
	if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
		raw := strings.TrimSuffix(strings.TrimPrefix(seg, "{"), "}")
		if idx := strings.Index(raw, ":"); idx != -1 {
			paramName := raw[:idx]
			pattern := raw[idx+1:]
			if pattern == "*" {
				return seg, nodeCatchAll, paramName, nil
			}
			if re, err := regexp.Compile("^" + pattern + "$"); err == nil {
				return seg, nodeParam, paramName, re
			}
			return seg, nodeParam, paramName, nil
		}
		return seg, nodeParam, raw, nil
	}
	if strings.HasPrefix(seg, ":") {
		return seg, nodeParam, seg[1:], nil
	}
	if strings.HasPrefix(seg, "*") {
		param := seg[1:]
		if param == "" {
			param = "*"
		}
		return seg, nodeCatchAll, param, nil
	}
	return seg, nodeStatic, "", nil
}

func splitPath(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return []string{}
	}
	return strings.Split(trimmed, "/")
}

// Insert adds a route pattern and handler into the Radix Trie.
func (t *Tree) Insert(method, pattern string, h http.Handler) {
	t.mu.Lock()
	defer t.mu.Unlock()

	segments := splitPath(pattern)
	curr := t.root

	for _, seg := range segments {
		raw, nType, paramName, re := parseSegment(seg)

		var child *trieNode
		for _, c := range curr.children {
			if c.nodeType == nType && c.segment == raw {
				child = c
				break
			}
		}

		if child == nil {
			child = newTrieNode(raw, nType, paramName, re)
			curr.children = append(curr.children, child)
		}
		curr = child

		if nType == nodeCatchAll {
			break
		}
	}

	curr.routes[strings.ToUpper(method)] = routeEntry{
		handler: h,
		pattern: pattern,
	}
}

// InsertMount mounts an http.Handler onto a prefix in the Radix Trie.
func (t *Tree) InsertMount(pattern string, handler http.Handler) {
	t.mu.Lock()
	defer t.mu.Unlock()

	cleanPrefix := "/" + strings.Trim(pattern, "/")
	if cleanPrefix == "/" {
		cleanPrefix = ""
	}

	segments := splitPath(pattern)
	curr := t.root

	for _, seg := range segments {
		raw, nType, paramName, re := parseSegment(seg)

		var child *trieNode
		for _, c := range curr.children {
			if c.nodeType == nType && c.segment == raw {
				child = c
				break
			}
		}

		if child == nil {
			child = newTrieNode(raw, nType, paramName, re)
			curr.children = append(curr.children, child)
		}
		curr = child
	}

	curr.mountServer = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, cleanPrefix)
		if p == "" || !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		r2 := new(http.Request)
		*r2 = *r
		urlCopy := *r.URL
		r2.URL = &urlCopy
		r2.URL.Path = p
		handler.ServeHTTP(w, r2)
	})
}

type matchResult struct {
	entry       *routeEntry
	params      map[string]string
	hasRoute    bool
	mountServer http.Handler
}

// Search queries the Radix Trie for a matching handler and extracts URL path parameters.
func (t *Tree) Search(method, path string) matchResult {
	t.mu.RLock()
	defer t.mu.RUnlock()

	segments := splitPath(path)
	params := make(map[string]string)

	return t.searchNode(t.root, segments, 0, strings.ToUpper(method), params)
}

func (t *Tree) searchNode(curr *trieNode, segments []string, idx int, method string, params map[string]string) matchResult {
	if curr.mountServer != nil {
		return matchResult{
			mountServer: curr.mountServer,
			params:      params,
			hasRoute:    true,
		}
	}

	if idx == len(segments) {
		if len(curr.routes) > 0 {
			if entry, ok := curr.routes[method]; ok {
				return matchResult{
					entry:    &entry,
					params:   params,
					hasRoute: true,
				}
			}
			return matchResult{
				hasRoute: true,
				params:   params,
			}
		}
		return matchResult{hasRoute: false}
	}

	seg := segments[idx]

	// 1. Try exact static match
	for _, child := range curr.children {
		if child.nodeType == nodeStatic && child.segment == seg {
			res := t.searchNode(child, segments, idx+1, method, params)
			if res.hasRoute {
				return res
			}
		}
	}

	// 2. Try param match with optional regex check
	for _, child := range curr.children {
		if child.nodeType == nodeParam {
			if child.regex != nil && !child.regex.MatchString(seg) {
				continue
			}
			params[child.paramName] = seg
			res := t.searchNode(child, segments, idx+1, method, params)
			if res.hasRoute {
				return res
			}
			delete(params, child.paramName)
		}
	}

	// 3. Try catch-all wildcard match
	for _, child := range curr.children {
		if child.nodeType == nodeCatchAll {
			remaining := strings.Join(segments[idx:], "/")
			params[child.paramName] = remaining
			if child.mountServer != nil {
				return matchResult{
					mountServer: child.mountServer,
					params:      params,
					hasRoute:    true,
				}
			}
			if len(child.routes) > 0 {
				if entry, ok := child.routes[method]; ok {
					return matchResult{
						entry:    &entry,
						params:   params,
						hasRoute: true,
					}
				}
				return matchResult{hasRoute: true, params: params}
			}
		}
	}

	return matchResult{hasRoute: false}
}
