package util

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"math/rand"
	"strings"
	"testing"
	"time"
)

func reverse(s string) (result string) {
	for _, v := range s {
		result = string(v) + result
	}
	return
}

func GenerateMappingString(seed []byte) (source, target, random string) {
	chars := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	tmpKey := sha512.Sum512(seed)
	total := int(tmpKey[0] + tmpKey[len(tmpKey)-1])
	for i := 0; i < total; i++ {
		if i > 50 {
			tmpKey = sha512.Sum512(tmpKey[0:64])
		} else {
			tmpKey = sha512.Sum512(tmpKey[0 : i+10])
		}
	}
	source = hex.EncodeToString(tmpKey[0:64])
	seedInt := int64(tmpKey[2] + tmpKey[0] + tmpKey[2] + tmpKey[1] + tmpKey[0] + tmpKey[7] + tmpKey[3] + tmpKey[0])
	rand.Seed(seedInt)
	rand.Shuffle(len(chars), func(i, j int) {
		chars[i], chars[j] = chars[j], chars[i]
	})

	source = string(chars)
	rand.Shuffle(len(chars), func(i, j int) {
		chars[i], chars[j] = chars[j], chars[i]
	})
	target = string(chars)

	rand.Seed(time.Now().UnixNano() + seedInt)
	rand.Shuffle(len(chars), func(i, j int) {
		chars[i], chars[j] = chars[j], chars[i]
	})
	random = string(chars)
	return
}

func translate(text string, sourceDict string, targetDict string) string {
	chars := []byte(text)
	for index, val := range chars {
		sourceIndex := strings.IndexByte(sourceDict, val)
		if sourceIndex >= 0 {
			chars[index] = targetDict[sourceIndex]
		}
	}
	return string(chars)
}

func SimpleEncrypt(data []byte, seed []byte) (ciphertext string, random string) {
	var source, target string
	plaintext := base64.StdEncoding.EncodeToString(data)
	source, target, random = GenerateMappingString(seed)
	random1 := random[0:10]
	random2 := random[10:20]
	random3 := random[20:30]
	random4 := random[30:40]
	random5 := random[40:50]
	random6 := random[50:62]
	realRandom := random3 + reverse(random1) + random6 + reverse(random4) + random2 + random5
	realSource := translate(source, realRandom, target)
	realTarget := translate(target, realRandom, source)
	ciphertext = translate(plaintext, realSource, realTarget)
	//fmt.Println("source", source)
	//fmt.Println("target", target)
	//fmt.Println("random", random)
	//fmt.Println("realRandom", realRandom)
	//fmt.Println("realSource", realSource)
	//fmt.Println("realTarget", realTarget)
	//fmt.Println("plaintext", plaintext)
	//fmt.Println("ciphertext", ciphertext)
	return
}

func TestSimpleEncrypt(t *testing.T) {
	plaintext := "测试中文"
	ciphertext, random := SimpleEncrypt([]byte(plaintext), []byte("20210731!@#$%)%^&%^&&*(&*"))
	t.Log(plaintext, ciphertext, random)
}
