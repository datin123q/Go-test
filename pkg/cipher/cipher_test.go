package cipher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncrypt(t *testing.T) {
	t.Run("valid key size", func(t *testing.T) {
		cipherKey := "CIPHERKEY01234567890123456789012"
		cipherIvKey := "CIPHERIVKEY01234"

		c, err := NewNiceCrypto(cipherKey, cipherIvKey)

		assert.NoError(t, err)
		assert.Equal(t, &niceCrypto{cipherKey, cipherIvKey}, c)
	})

	t.Run("invalid key size", func(t *testing.T) {
		cipherKey := "CIPHERKEY01234567"
		cipherIvKey := "CIPHERIVKEY"

		_, err := NewNiceCrypto(cipherKey, cipherIvKey)

		assert.Error(t, err)
	})

	t.Run("plain test `Abel Ko` should be equal `HD4PbbIOrV10Qc/+/7e+IA==`", func(t *testing.T) {
		cipherKey := "CIPHERKEY01234567890123456789012"
		cipherIvKey := "CIPHERIVKEY01234"
		plainText := "Abel Ko"
		expected := "kOYY6cToDztsoc7RWZl6vA=="

		c, err := NewNiceCrypto(cipherKey, cipherIvKey)
		assert.NoError(t, err)

		cipher, err := c.Encrypt(plainText)
		assert.NoError(t, err)
		assert.EqualValues(t, expected, cipher)
	})
}

func TestDecrypt(t *testing.T) {
	t.Run("decrypted text `bgyE7m4EeLqQvtKnihJMfg==` should be equal `Abel Ko`", func(t *testing.T) {
		cipherKey := "CIPHERKEY01234567890123456789012"
		cipherIvKey := "CIPHERIVKEY01234"
		chiperText := "kOYY6cToDztsoc7RWZl6vA=="
		expected := "Abel Ko"

		c, err := NewNiceCrypto(cipherKey, cipherIvKey)
		assert.NoError(t, err)

		actual, err := c.Decrypt(chiperText)
		assert.NoError(t, err)
		assert.EqualValues(t, expected, actual)
	})
}
