package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/nacl/box"
	"net/http"
	"os"
)

type createSecretRequest struct {
	Secret     string
	SecretName string
	RepoOwner  string
	RepoName   string
}

type encodeSecretRequest struct {
	PublicKey string
	Secret    string
}

type ghPublicKeyResponse struct {
	Key   string `json:"key,omitempty"`
	KeyId string `json:"key_id,omitempty"`
}

type ghCreateSecretRequest struct {
	EncryptedValue string `json:"encrypted_value,omitempty"`
	KeyId          string `json:"key_id,omitempty"`
}

func createSecret(c *gin.Context) {
	var request createSecretRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(400, gin.H{})
		return
	}

	client := resty.
		New().
		SetAuthToken(os.Getenv("GITHUB_TOKEN")).
		SetBaseURL(os.Getenv("GITHUB_URL"))

	var publicKey ghPublicKeyResponse
	resp, err := client.R().
		SetResult(&publicKey).
		SetPathParams(map[string]string{
			"repoOwner":  request.RepoOwner,
			"repoName":   request.RepoName,
			"secretName": request.SecretName,
		}).
		Get("/repos/{repoOwner}/{repoName}/actions/secrets/public-key")
	if err != nil || resp.StatusCode() != http.StatusOK {
		handleGHError(c, err, resp)
		return
	}

	encoded, err := encodeWithPublicKey(request.Secret, publicKey.Key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	resp, err = client.R().
		SetBody(ghCreateSecretRequest{
			EncryptedValue: encoded,
			KeyId:          publicKey.KeyId,
		}).
		SetPathParams(map[string]string{
			"repoOwner":  request.RepoOwner,
			"repoName":   request.RepoName,
			"secretName": request.SecretName,
		}).
		Put("/repos/{repoOwner}/{repoName}/actions/secrets/{secretName}")

	if err != nil || !(resp.StatusCode() == http.StatusCreated || resp.StatusCode() == http.StatusNoContent) {
		handleGHError(c, err, resp)
		return
	}

	c.JSON(http.StatusCreated, gin.H{})
}

func handleGHError(c *gin.Context, err error, resp *resty.Response) {
	if err == nil {
		err = fmt.Errorf("error response: %s, with code: %d", resp, resp.StatusCode())
	}
	c.JSON(http.StatusInternalServerError, err)

	log.Error(err)
	return
}

func encodeSecret(c *gin.Context) {
	var request encodeSecretRequest

	if err := c.BindJSON(&request); err != nil {
		c.JSON(400, gin.H{})
	}

	encoded, err := encodeWithPublicKey(request.Secret, request.PublicKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"encodedSecret": encoded,
	})
}

func encodeWithPublicKey(text string, publicKey string) (string, error) {
	// Decode the public key from base64
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return "", err
	}

	// Decode the public key
	var publicKeyDecoded [32]byte
	copy(publicKeyDecoded[:], publicKeyBytes)

	// Encrypt the secret value
	encrypted, err := box.SealAnonymous(nil, []byte(text), (*[32]byte)(publicKeyBytes), rand.Reader)

	if err != nil {
		return "", err
	}
	// Encode the encrypted value in base64
	encryptedBase64 := base64.StdEncoding.EncodeToString(encrypted)

	return encryptedBase64, nil
}
