package util

import "fmt"

// UnsealKeyID is the ID that used as key name when storing unseal key
func UnsealKeyID(prefix string, i int) string {
	return fmt.Sprintf("%s-unseal-key-%d", prefix, i)
}

// RootTokenID is the ID that used as key name when storing root token
func RootTokenID(prefix string) string {
	return fmt.Sprintf("%s-root-token", prefix)
}
