package utils

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
)

/*
Restituisce la sottostringa tramite stringa di inizio e fine specificate
*/
func GetStringInBetween(str string, startS string, endS string) (result string) {
	s := strings.Index(str, startS)
	if s == -1 {
		return result
	}
	newS := str[s+len(startS):]
	e := strings.Index(newS, endS)
	if e == -1 {
		return result
	}
	result = newS[:e]
	return result
}

/*
Indica se una sottostringa è presente o meno all'interno di una Slice
*/
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

/*
Rimuove un elemento da una slice
*/
func RemoveElement(slice []string, remove string) []string {
	var i int
	for i = 0; i < len(slice); i++ {
		if slice[i] == remove {
			break
		}
	}

	slice[i] = slice[len(slice)-1]
	return slice[:len(slice)-1]
}

/*
Elimina tutti i file contenuti in una cartella specificata
*/
func ClearDir(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return err
	}
	for _, file := range files {
		err = os.RemoveAll(file)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
Effettua l'hashing sha256 di una stringa
*/
func HashString(str string) [32]byte {
	return sha256.Sum256([]byte(str))
}
