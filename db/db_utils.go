package db

import "go.mongodb.org/mongo-driver/v2/bson"

func ToBSONDocument(v any) (doc bson.D, err error) {
	data, err := bson.Marshal(v)
	if err != nil {
		return nil, err
	}

	err = bson.Unmarshal(data, &doc)
	if err != nil {
		return nil, err
	}

	return
}
