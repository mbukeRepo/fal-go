package fal

const REST_API_URL = "https://rest.alpha.fal.ai"

type IStorage interface {
	Upload(fileData []byte) string
	TransformInput(inputData []string) string
	InitiateUpload(fileData []byte) string
}

type InitiateUploadResult struct {
	FileUrl   string
	UploadUrl string
}

type InitiateUploadData struct {
	FileName    string
	ContentType *string
}

type Storage struct{}

func (s *Storage) InitiateUpload(fileData []byte) {

}

func (s *Storage) Upload(fileData []byte) string {
	return ""
}

func (s *Storage) TransformInput(inputData []string) string {
	return ""
}
