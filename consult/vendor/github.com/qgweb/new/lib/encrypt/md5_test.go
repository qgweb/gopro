package encrypt

import (
	"testing"
)

func TestMd5_1(t *testing.T) {
	if GetEncoder(TYPE_MD5).Encode("1234") == "81dc9bdb52d04dc20036dbd8313ed055" {
		t.Log("用力1测试通过")
	} else {
		t.Log("用例1测试不通过")
	}
}

func TestMd5_2(t *testing.T) {
	if GetEncoder(TYPE_MD5).Encode("12345") == "81dc9bdb52d04dc20036dbd8313ed055" {
		t.Log("用力2测试通过")
	} else {
		t.Error("用例2测试不通过")
	}
}

func BenchmarkAdd(t *testing.B) {
	for i := 0; i < 1000; i++ {
		GetEncoder(TYPE_MD5).Encode("1234")
	}
}
