package server

const (
	HistoryTypeID                int = 963985
	HistoryDeleteDirectiveTypeID int = 150251
)

type SyncEntity struct {
	ClientID               string
	ID                     string
	ParentID               *string `dynamodbav:",omitempty"`
	Version                *int64
	Mtime                  *int64
	Ctime                  *int64
	Name                   *string `dynamodbav:",omitempty"`
	NonUniqueName          *string `dynamodbav:",omitempty"`
	ServerDefinedUniqueTag *string `dynamodbav:",omitempty"`
	Deleted                *bool
	OriginatorCacheGUID    *string `dynamodbav:",omitempty"`
	OriginatorClientItemID *string `dynamodbav:",omitempty"`
	DataType               *int
	Folder                 *bool
	ClientDefinedUniqueTag *string `dynamodbav:",omitempty"`
	UniquePosition         []byte  `dynamodbav:",omitempty"`
	DataTypeMtime          *string
	ExpirationTime         *int64
}
