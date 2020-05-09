package utils_test

import (
	"github.com/overtalk/qnet/utils"
	"testing"
)

func TestHashCode(t *testing.T) {
	hashCode := utils.HashCode("xxx")
	t.Log(hashCode)
}

func BenchmarkHashCode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		utils.HashCode("asdfasdfasdfasdfasdfasdfasdfasd")
	}
}
