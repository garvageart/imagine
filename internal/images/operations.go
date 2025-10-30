package images

import (
	"encoding/base64"
	"fmt"
	"image"
	"os"
)

func EncodeThumbhashToString(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func DecodeThumbhashString(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

func CreateImageDir(uid string) error {
	return os.MkdirAll(fmt.Sprintf("%s/%s", Directory, uid), os.ModePerm)
}

func DeleteImageDir(uid string) error {
	return os.RemoveAll(fmt.Sprintf("%s/%s", Directory, uid))
}

func GetImagePath(uid, fileName string) string {
	return fmt.Sprintf("%s/%s/%s", Directory, uid, fileName)
}

func ReadImage(uid, fileName string) ([]byte, error) {
	return os.ReadFile(GetImagePath(uid, fileName))
}

func ReadFileAsGoImage(uid, fileName string) (imageData image.Image, format string, err error) {
	file, err := os.Open(GetImagePath(uid, fileName))
	if err != nil {
		return nil, "", err
	}

	defer file.Close()

	imageData, format, err = image.Decode(file)
	if err != nil {
		return nil, format, err
	}
	return imageData, format, nil
}

func SaveImage(data []byte, uid, fileName string) error {
	err := CreateImageDir(uid)
	if err != nil {
		return err
	}

	imagesDir := GetImagePath(uid, fileName)
	return os.WriteFile(imagesDir, data, 0644)
}
