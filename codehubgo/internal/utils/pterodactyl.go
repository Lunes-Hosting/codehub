package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	
	"codehubgo/internal/config"
)

// EggVariable represents a variable from a Pterodactyl egg
type EggVariable struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	EnvVariable string `json:"env_variable"`
	DefaultValue string `json:"default_value"`
	Required    bool   `json:"required"`
}

// Egg represents a Pterodactyl egg
type Egg struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// GetAvailableEggs fetches available eggs from Pterodactyl API
func GetAvailableEggs() ([]Egg, error) {
	url := fmt.Sprintf("%s/api/application/nests/%s/eggs", config.PterodactylURL, config.AutodeployNestID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "Bearer "+config.PterodactylAdminKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get eggs: %s", resp.Status)
	}
	
	var response struct {
		Data []struct {
			Attributes Egg `json:"attributes"`
		} `json:"data"`
	}
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	
	eggs := make([]Egg, len(response.Data))
	for i, egg := range response.Data {
		eggs[i] = egg.Attributes
	}
	
	return eggs, nil
}

// GetEggVariables fetches variables for a specific egg from Pterodactyl API
func GetEggVariables(eggID string) ([]EggVariable, error) {
	url := fmt.Sprintf("%s/api/application/nests/%s/eggs/%s", config.PterodactylURL, config.AutodeployNestID, eggID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "Bearer "+config.PterodactylAdminKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get egg variables: %s", resp.Status)
	}
	
	var response struct {
		Attributes struct {
			Variables []EggVariable `json:"variables"`
		} `json:"attributes"`
	}
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	
	return response.Attributes.Variables, nil
}

// CreateServer creates a new server in Pterodactyl
func CreateServer(name, ownerEmail string, eggID int, environment map[string]string) (int, error) {
	url := fmt.Sprintf("%s/api/application/servers", config.PterodactylURL)
	
	// Convert environment map to array of objects
	var envVars []map[string]string
	for key, value := range environment {
		envVars = append(envVars, map[string]string{
			"key":   key,
			"value": value,
		})
	}
	
	// Create request payload
	payload := map[string]interface{}{
		"name":        name,
		"user":        ownerEmail,
		"egg":         eggID,
		"docker_image": "ghcr.io/pterodactyl/yolks:nodejs_18",  // Default image, adjust as needed
		"startup":     "npm start",
		"environment": envVars,
		"limits": map[string]interface{}{
			"memory": 1024,
			"swap":   0,
			"disk":   1024,
			"io":     500,
			"cpu":    100,
		},
		"feature_limits": map[string]interface{}{
			"databases":   1,
			"backups":     1,
			"allocations": 1,
		},
		"allocation": map[string]interface{}{
			"default": 1,
		},
	}
	
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return 0, err
	}
	
	req.Header.Set("Authorization", "Bearer "+config.PterodactylAdminKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return 0, fmt.Errorf("failed to create server: %s - %s", resp.Status, string(body))
	}
	
	var response struct {
		Attributes struct {
			ID int `json:"id"`
		} `json:"attributes"`
	}
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}
	
	return response.Attributes.ID, nil
}

// DeleteServer deletes a server from Pterodactyl
func DeleteServer(serverID int) error {
	url := fmt.Sprintf("%s/api/application/servers/%d", config.PterodactylURL, serverID)
	
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	
	req.Header.Set("Authorization", "Bearer "+config.PterodactylAdminKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusNoContent {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete server: %s - %s", resp.Status, string(body))
	}
	
	return nil
}
