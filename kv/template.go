package kv

import (
	"encoding/json"
	"fmt"
	"strings"
)

const TemplatePrefix = "mg_templates:"

type TemplateItem struct {
	Slug     string `json:"slug"`
	Workflow string `json:"workflow"`
}

func templateStringConcat(templateString string) string {
	return fmt.Sprintf("mg_templates:%s", templateString)
}

type TemplateStoreManager struct {
	store *Store
}

func NewTemplateStoreManager(store *Store) *TemplateStoreManager {
	return &TemplateStoreManager{store: store}
}

func (ts *TemplateStoreManager) CreateTemplate(template TemplateItem) error {
	key := templateStringConcat(template.Slug)
	return ts.store.Set(key, template)
}

func (ts *TemplateStoreManager) GetTemplate(slug string) (*TemplateItem, error) {
	var template TemplateItem
	key := templateStringConcat(slug)
	err := ts.store.Get(key, &template)
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (ts *TemplateStoreManager) UpdateTemplate(template TemplateItem) error {
	return ts.CreateTemplate(template)
}

func (ts *TemplateStoreManager) DeleteTemplate(slug string) error {
	existing, err := ts.GetTemplate(slug)
	if err != nil {
		return fmt.Errorf("template not found: %v", err)
	}
	if existing == nil {
		return fmt.Errorf("template '%s' does not exist", slug)
	}

	return ts.store.Delete(templateStringConcat(slug))
}

func (ts *TemplateStoreManager) ListTemplates() ([]TemplateItem, error) {
	// Use the same prefix as used in templateStringConcat
	keys, err := ts.store.List(TemplatePrefix)
	if err != nil {
		return nil, err
	}

	templates := make([]TemplateItem, 0, len(keys))
	for _, key := range keys {
		var template TemplateItem
		err := ts.store.Get(key, &template)
		if err != nil {
			return nil, err
		}
		templates = append(templates, template)
	}

	return templates, nil
}

func (ts *TemplateStoreManager) SearchTemplates(query string) ([]TemplateItem, error) {
	allTemplates, err := ts.ListTemplates()
	if err != nil {
		return nil, err
	}

	var results []TemplateItem
	for _, template := range allTemplates {
		if strings.Contains(strings.ToLower(template.Slug), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(template.Workflow), strings.ToLower(query)) {
			results = append(results, template)
		}
	}

	return results, nil
}

func (ts *TemplateStoreManager) ExportTemplate(slug string) (string, error) {
	template, err := ts.GetTemplate(slug)
	if err != nil {
		return "", err
	}

	jsonData, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func (ts *TemplateStoreManager) ImportTemplate(jsonData string) error {
	var template TemplateItem
	err := json.Unmarshal([]byte(jsonData), &template)
	if err != nil {
		return err
	}

	return ts.CreateTemplate(template)
}
