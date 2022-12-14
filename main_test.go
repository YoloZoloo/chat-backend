package main

import (
	"bytes"
	"chat-backend/src"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	. "time"
)

func TestCurlMainScreen(t *testing.T) {
	go main()
	Sleep(3 * Second)
	requestURL := "http://localhost:9999"
	res, err := http.Get(requestURL)
	if err != nil {
		t.Errorf("error: %q", err)
	}
	//must return 405
	if res.StatusCode != 405 {
		t.Errorf("resp status: %d", res.StatusCode)
	}
}

func TestLoginTest(t *testing.T) {
	requestURL := "http://localhost:9999/api/login"
	userName := os.Getenv("LOGIN_TEST_USER_ID")
	pass := os.Getenv("LOGIN_TEST_USER_PASS")
	creds := &src.UserCredentials{UserName: userName, Password: pass}
	t.Log("data: ", creds)
	data, _ := json.Marshal(creds)
	t.Log("data: ", data)
	res, err := http.Post(requestURL, "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		t.Errorf("error: %q", err)
	}
	if res.StatusCode != 200 {
		t.Errorf("resp status: %d", res.StatusCode)
	}
}
