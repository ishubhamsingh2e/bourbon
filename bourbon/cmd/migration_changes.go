package cmd

// MigrationChanges represents all types of changes detected
type MigrationChanges struct {
	NewModels      []ModelInfo            // Completely new models
	DeletedModels  []string               // Model names that were deleted
	NewFields      map[string][]FieldInfo // modelName -> new fields
	DeletedFields  map[string][]FieldInfo // modelName -> deleted fields
	ModifiedFields map[string][]FieldInfo // modelName -> modified fields (type or tag changed)
}

// DetectAllChanges performs comprehensive change detection
func DetectAllChanges(appName string, currentModels []ModelInfo) (*MigrationChanges, error) {
	changes := &MigrationChanges{
		NewModels:      []ModelInfo{},
		DeletedModels:  []string{},
		NewFields:      make(map[string][]FieldInfo),
		DeletedFields:  make(map[string][]FieldInfo),
		ModifiedFields: make(map[string][]FieldInfo),
	}

	state, err := LoadMigrationState()
	if err != nil {
		return nil, err
	}

	// If no previous state, everything is new
	if state.Apps[appName] == nil {
		changes.NewModels = currentModels
		return changes, nil
	}

	// Create lookups
	currentModelMap := make(map[string]*ModelInfo)
	for i := range currentModels {
		currentModelMap[currentModels[i].Name] = &currentModels[i]
	}

	storedModels := state.Apps[appName].Models

	// Detect new and modified models
	for _, current := range currentModels {
		stored, exists := storedModels[current.Name]
		if !exists {
			// Completely new model
			changes.NewModels = append(changes.NewModels, current)
			continue
		}

		// Model exists, check fields
		storedFieldMap := make(map[string]FieldState)
		for _, f := range stored.Fields {
			storedFieldMap[f.Name] = f
		}

		currentFieldMap := make(map[string]FieldInfo)
		for _, f := range current.Fields {
			currentFieldMap[f.Name] = f
		}

		// Detect new and modified fields
		for _, currentField := range current.Fields {
			storedField, exists := storedFieldMap[currentField.Name]
			if !exists {
				// New field
				changes.NewFields[current.Name] = append(changes.NewFields[current.Name], currentField)
			} else if storedField.Type != currentField.Type || storedField.Tag != currentField.Tag {
				// Modified field
				changes.ModifiedFields[current.Name] = append(changes.ModifiedFields[current.Name], currentField)
			}
		}

		// Detect deleted fields
		for _, storedField := range stored.Fields {
			if _, exists := currentFieldMap[storedField.Name]; !exists {
				// Deleted field
				changes.DeletedFields[current.Name] = append(changes.DeletedFields[current.Name], FieldInfo{
					Name: storedField.Name,
					Type: storedField.Type,
					Tag:  storedField.Tag,
				})
			}
		}
	}

	// Detect deleted models
	for modelName := range storedModels {
		if _, exists := currentModelMap[modelName]; !exists {
			changes.DeletedModels = append(changes.DeletedModels, modelName)
		}
	}

	return changes, nil
}

// HasChanges returns true if there are any changes
func (c *MigrationChanges) HasChanges() bool {
	return len(c.NewModels) > 0 ||
		len(c.DeletedModels) > 0 ||
		len(c.NewFields) > 0 ||
		len(c.DeletedFields) > 0 ||
		len(c.ModifiedFields) > 0
}

// HasDestructiveChanges returns true if there are any destructive changes
func (c *MigrationChanges) HasDestructiveChanges() bool {
	return len(c.DeletedModels) > 0 || len(c.DeletedFields) > 0
}
