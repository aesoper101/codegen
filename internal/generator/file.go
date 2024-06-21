package generator

type File struct {
	Name          string
	Content       string
	IsNew         bool
	BackupContent string
}

func (f *File) Clone() *File {
	return &File{
		Name:          f.Name,
		Content:       f.Content,
		IsNew:         f.IsNew,
		BackupContent: f.BackupContent,
	}
}
