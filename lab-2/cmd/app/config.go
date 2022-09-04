package main

import (
	"github.com/minio/pkg/env"
	"strings"
)

func GetAdminPanel() string {
	return env.Get("DVKA_LAB2_ADMIN_PANEL", "off")
}

func GetFlag() string {
	return env.Get("DVKA_LAB2_FLAG", "flag{}")
}

func GetPodIP() string {
	return env.Get("POD_IP", "10.0.0.1")
}

type RunCommandRequest struct {
	Command  string `json:"command,omitempty"`
	Hostname string `json:"hostname,omitempty"`
}

func IsIPv6(address string) bool {
	return strings.Count(address, ":") >= 2
}
