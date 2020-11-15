package controllers

import (
	"context"
	"os"
	"strings"
	"time"

	gocloak "github.com/Nerzal/gocloak/v7"
	"k8s.io/klog"
)

// CheckAndRenewToken checks if the token is expired, if so it renews it
func CheckAndRenewToken(ctx context.Context, kcClient gocloak.GoCloak, token **gocloak.JWT) error {

	_, claims, err := kcClient.DecodeAccessToken(ctx, (*token).AccessToken, "master", "")
	if err != nil {
		klog.Error(err, "problems when decoding token")
		return err
	}
	tokenExpiredDiff := time.Unix(int64((*claims)["exp"].(float64)), 0).Sub(time.Now()).Seconds()

	if tokenExpiredDiff < 60 {
		kcAdminUser := os.Getenv("KEYCLOAK_ADMIN_USER")
		kcAdminPsw := os.Getenv("KEYCLOAK_ADMIN_PSW")
		*token, err = kcClient.LoginAdmin(ctx, kcAdminUser, kcAdminPsw, "master")
		if err != nil {
			klog.Error(err, "Error when renewing token")
			return err
		}
		klog.Info("Token renewed successfully")
		return nil
	}
	klog.Info("No need to renew token")
	return nil
}

// GetClientID returns the ID of the target client given the human id, to be used with the gocloak library
func GetClientID(ctx context.Context, kcClient gocloak.GoCloak, token string, realmName string, targetClient string) (string, error) {
	var targetClientID string

	clients, err := kcClient.GetClients(ctx, token, realmName, gocloak.GetClientsParams{ClientID: &targetClient})
	if err != nil {
		klog.Error(err, "Error when getting k8s client")
		return "", err
	} else if len(clients) > 1 {
		klog.Error(nil, "too many k8s clients")
		return "", err
	} else if len(clients) < 0 {
		klog.Error(nil, "no k8s client")
		return "", err

	} else {
		targetClientID = *clients[0].ID
		klog.Info("Got client id", "id", targetClientID)
		return targetClientID, nil
	}

}

func createKcRole(ctx context.Context, kcClient gocloak.GoCloak, token string, realmName string, targetClientID string, newRoleName string) error {
	// check if keycloak role already esists

	_, err := kcClient.GetClientRole(ctx, token, realmName, targetClientID, newRoleName)
	if err != nil && strings.Contains(err.Error(), "404 Not Found: Could not find role") {
		// error corresponds to "not found"
		// need to create new role
		klog.Info("Role didn't exists", "role", newRoleName)
		tr := true
		createdRoleName, err := kcClient.CreateClientRole(ctx, token, "crownlabs", targetClientID, gocloak.Role{Name: &newRoleName, ClientRole: &tr})
		if err != nil {
			klog.Error(err, "Error when creating role")
			return err
		}
		klog.Info("Role created", "rolename", createdRoleName)
		return nil
	} else if err != nil {
		klog.Error(err, "Error when getting user role")
		return err
	} else {
		klog.Info("Role already existed", "role", newRoleName)
		return nil
	}
}
