***REMOVED***

***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***

***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***

***REMOVED***
***REMOVED***

func (db DB***REMOVED*** Connect(***REMOVED*** (*mongo.Client, error***REMOVED*** {

	host := fmt.Sprintf("%s:%d", db.Address, db.Port***REMOVED***
	fmt.Println("Connecting to MongoDB..."***REMOVED***

***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED******REMOVED***

***REMOVED***
		fmt.Println(err***REMOVED***
***REMOVED***
***REMOVED***

	if err := client.Ping(db.Context, nil***REMOVED***; err != nil {
		fmt.Println(err***REMOVED***
***REMOVED***
***REMOVED***

	fmt.Println("Pinged your deployment. You successfully connected to MongoDB at", db.Address, "on port", db.Port***REMOVED***

	fmt.Println("Using database", db.DatabaseName***REMOVED***

***REMOVED***
***REMOVED***

func (db DB***REMOVED*** Disconnect(client *mongo.Client***REMOVED*** error {
	err := client.Disconnect(db.Context***REMOVED***

***REMOVED***
***REMOVED***
***REMOVED***

***REMOVED***
***REMOVED***

func (db DB***REMOVED*** Delete(document bson.D***REMOVED*** (*mongo.DeleteResult, error***REMOVED*** {
	result, err := db.Database.Collection(db.Collection***REMOVED***.DeleteOne(db.Context, document***REMOVED***

***REMOVED***
	***REMOVED***, err
***REMOVED***

***REMOVED***
***REMOVED***

func (db DB***REMOVED*** Exists(filter bson.D***REMOVED*** (bool, error***REMOVED*** {
	err := db.Database.Collection(db.Collection***REMOVED***.FindOne(db.Context, filter***REMOVED***.Err(***REMOVED***

***REMOVED***
***REMOVED***
***REMOVED***
	***REMOVED***

		fmt.Println(err***REMOVED***
***REMOVED***
***REMOVED***

***REMOVED***
***REMOVED***

func (db DB***REMOVED*** Find(filter bson.D, result any***REMOVED*** (*mongo.Cursor, error***REMOVED*** {
	cursor, err := db.Database.Collection(db.Collection***REMOVED***.Find(db.Context, filter***REMOVED***

***REMOVED***
	***REMOVED***, err
***REMOVED***

	cursorErr := cursor.All(db.Context, result***REMOVED***

***REMOVED***
		fmt.Println(cursorErr***REMOVED***
	***REMOVED***, cursorErr
***REMOVED***

***REMOVED***
***REMOVED***

func (db DB***REMOVED*** FindOne(filter bson.D, result any***REMOVED*** {
	db.Database.Collection(db.Collection***REMOVED***.FindOne(db.Context, filter***REMOVED***.Decode(result***REMOVED***
***REMOVED***

func (db DB***REMOVED*** Insert(document bson.D***REMOVED*** (*mongo.InsertOneResult, error***REMOVED*** {
	result, err := db.Database.Collection(db.Collection***REMOVED***.InsertOne(db.Context, document***REMOVED***

***REMOVED***
	***REMOVED***, err
***REMOVED***

***REMOVED***
***REMOVED***

func (db DB***REMOVED*** Update(filter bson.D, document bson.D***REMOVED*** (*mongo.UpdateResult, error***REMOVED*** {
	result, err := db.Database.Collection(db.Collection***REMOVED***.UpdateOne(db.Context, filter, document***REMOVED***

***REMOVED***
	***REMOVED***, err
***REMOVED***

***REMOVED***
***REMOVED***

func (db DB***REMOVED*** UpdateMany(filter bson.D, documents []bson.D***REMOVED*** (*mongo.UpdateResult, error***REMOVED*** {
	result, err := db.Database.Collection(db.Collection***REMOVED***.UpdateMany(db.Context, filter, documents***REMOVED***

***REMOVED***
	***REMOVED***, err
***REMOVED***

***REMOVED***
***REMOVED***

func (db DB***REMOVED*** DeleteMany(documents []bson.D***REMOVED*** (*mongo.DeleteResult, error***REMOVED*** {
	result, err := db.Database.Collection(db.Collection***REMOVED***.DeleteMany(db.Context, documents***REMOVED***

***REMOVED***
	***REMOVED***, err
***REMOVED***

***REMOVED***
***REMOVED***

func (db DB***REMOVED*** InsertMany(documents []bson.D***REMOVED*** (*mongo.InsertManyResult, error***REMOVED*** {
	result, err := db.Database.Collection(db.Collection***REMOVED***.InsertMany(db.Context, documents***REMOVED***

***REMOVED***
	***REMOVED***, err
***REMOVED***

***REMOVED***
***REMOVED***

func (db DB***REMOVED*** ReplaceOne(filter bson.D, replacement bson.D***REMOVED*** (*mongo.UpdateResult, error***REMOVED*** {
	result, err := db.Database.Collection(db.Collection***REMOVED***.ReplaceOne(db.Context, filter, replacement***REMOVED***

***REMOVED***
	***REMOVED***, err
***REMOVED***

***REMOVED***
***REMOVED***

func Initis(***REMOVED*** error {
	mongoCtx, cancelMongo := context.WithTimeout(context.Background(***REMOVED***, 60*time.Second***REMOVED***
	defer cancelMongo(***REMOVED***
***REMOVED***

***REMOVED***
***REMOVED***
***REMOVED***

***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***
		User:         os.Getenv("MONGO_USER"***REMOVED***,
		Password:     os.Getenv("MONGO_PASSWORD"***REMOVED***,
***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***
***REMOVED***

	client, err := db.Connect(***REMOVED***
	defer func(***REMOVED*** {
		db.Disconnect(client***REMOVED***
		fmt.Println("Disconnected from MongoDB"***REMOVED***
***REMOVED***(***REMOVED***

***REMOVED***
		fmt.Println("error connecting mongo db"***REMOVED***
***REMOVED***
***REMOVED***

***REMOVED***
***REMOVED***
