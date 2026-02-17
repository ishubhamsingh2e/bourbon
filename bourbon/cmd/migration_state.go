package cmd

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type FieldState struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Tag  string `json:"tag"`
}

type ModelState struct {
	Name   string       `json:"name"`
	Hash   string       `json:"hash"`
	Fields []FieldState `json:"fields"`
}

type AppMigrationState struct {
	LastHash      string                 `json:"last_hash"`
	LastMigration string                 `json:"last_migration"`
	Models        map[string]*ModelState `json:"models"` // model name -> state
}

type MigrationState struct {
	Apps map[string]*AppMigrationState `json:"apps"`
}

func getStateFilePath() string {
	return filepath.Join(".bourbon", "migration_state.json")
}

func LoadMigrationState() (*MigrationState, error) {
	statePath := getStateFilePath()

	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return &MigrationState{
			Apps: make(map[string]*AppMigrationState),
		}, nil
	}

	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state MigrationState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	if state.Apps == nil {
		state.Apps = make(map[string]*AppMigrationState)
	}

	return &state, nil
}

func SaveMigrationState(state *MigrationState) error {
	stateDir := filepath.Dir(getStateFilePath())
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create .bourbon directory: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(getStateFilePath(), data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

func ComputeModelsHash(models []ModelInfo) string {
	h := sha256.New()

	// Sort models by name for consistent hashing
	sortedModels := make([]ModelInfo, len(models))
	copy(sortedModels, models)
	sort.Slice(sortedModels, func(i, j int) bool {
		return sortedModels[i].Name < sortedModels[j].Name
	})

	for _, model := range sortedModels {
		// Hash: ModelName
		h.Write([]byte(model.Name))

		// Sort fields for consistent hashing
		sortedFields := make([]FieldInfo, len(model.Fields))
		copy(sortedFields, model.Fields)
		sort.Slice(sortedFields, func(i, j int) bool {
			return sortedFields[i].Name < sortedFields[j].Name
		})

		for _, field := range sortedFields {
			// Hash: FieldName:FieldType:Tag
			h.Write([]byte(fmt.Sprintf("%s:%s:%s", field.Name, field.Type, field.Tag)))
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func ComputeSingleModelHash(model ModelInfo) string {
	h := sha256.New()
	h.Write([]byte(model.Name))

	// Sort fields for consistent hashing
	sortedFields := make([]FieldInfo, len(model.Fields))
	copy(sortedFields, model.Fields)
	sort.Slice(sortedFields, func(i, j int) bool {
		return sortedFields[i].Name < sortedFields[j].Name
	})

	for _, field := range sortedFields {
		h.Write([]byte(fmt.Sprintf("%s:%s:%s", field.Name, field.Type, field.Tag)))
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func DetectModelChanges(appName string, models []ModelInfo) (bool, error) {
	state, err := LoadMigrationState()
	if err != nil {
		return false, err
	}

	currentHash := ComputeModelsHash(models)

	appState, exists := state.Apps[appName]
	if !exists || appState.LastHash != currentHash {
		return true, nil
	}

	return false, nil
}

func UpdateMigrationState(appName string, models []ModelInfo, migrationID string) error {
	state, err := LoadMigrationState()
	if err != nil {
		return err
	}

	if state.Apps[appName] == nil {
		state.Apps[appName] = &AppMigrationState{
			Models: make(map[string]*ModelState),
		}
	}

	state.Apps[appName].LastHash = ComputeModelsHash(models)
	state.Apps[appName].LastMigration = migrationID

	// Update individual model states with field information
	for _, model := range models {
		fields := make([]FieldState, len(model.Fields))
		for i, f := range model.Fields {
			fields[i] = FieldState{
				Name: f.Name,
				Type: f.Type,
				Tag:  f.Tag,
			}
		}
		
		state.Apps[appName].Models[model.Name] = &ModelState{
			Name:   model.Name,
			Hash:   ComputeSingleModelHash(model),
			Fields: fields,
		}
	}

	return SaveMigrationState(state)
}

// DetectDeletedFields compares current models with stored state to find deleted fields
func DetectDeletedFields(appName string, models []ModelInfo) map[string][]string {
	deletedFields := make(map[string][]string) // modelName -> []fieldName
	
	state, err := LoadMigrationState()
	if err != nil || state.Apps[appName] == nil {
		return deletedFields
	}
	
	// Create lookup for current models
	currentModels := make(map[string]*ModelInfo)
	for i := range models {
		currentModels[models[i].Name] = &models[i]
	}
	
	// Check each stored model
	for modelName, storedModel := range state.Apps[appName].Models {
		currentModel, exists := currentModels[modelName]
		if !exists {
			// Entire model deleted - we'll handle this separately
			continue
		}
		
		// Create lookup for current fields
		currentFields := make(map[string]bool)
		for _, f := range currentModel.Fields {
			currentFields[f.Name] = true
		}
		
		// Find deleted fields
		for _, storedField := range storedModel.Fields {
			if !currentFields[storedField.Name] {
				deletedFields[modelName] = append(deletedFields[modelName], storedField.Name)
			}
		}
	}
	
	return deletedFields
}
