package controller

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/h2non/filetype/types"
	"github.com/hyperjiang/gallery-service/app/model"
	"github.com/hyperjiang/gallery-service/app/provider"
)

// FileController - file controller
type FileController struct{}

// Form - show upload form
func (ctrl *FileController) Form(c *gin.Context) {
	c.HTML(http.StatusOK, "upload.html", nil)
}

// Upload - save upload file
func (ctrl *FileController) Upload(c *gin.Context) {

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	mime := types.NewMIME(file.Header.Get("Content-Type"))

	// check if the uploaded file is image or video
	reg := regexp.MustCompile("^(.+)\\/.+$")
	matches := reg.FindStringSubmatch(mime.Value)
	if len(matches) != 2 || (matches[1] != "image" && matches[1] != "video") {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Unsupported file type: " + mime.Value,
		})
		return
	}

	// open the file and calculate the md5 checksum of the file
	f, err := file.Open()
	defer f.Close()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	checksum := fmt.Sprintf("%x", h.Sum(nil))

	// make the destination location
	dir := filepath.Join(
		provider.DI().Config().Server.UploadDir,
		checksum[0:2],
		checksum[2:4],
	)
	filename := checksum + "." + mime.Subtype
	dst := filepath.Join(dir, filename)

	// upload the file to specific dst.
	os.MkdirAll(dir, os.ModePerm)
	c.SaveUploadedFile(file, dst)

	fileModel := &model.File{
		Name:     file.Filename,
		Path:     dst,
		Type:     mime.Value,
		Size:     uint32(file.Size),
		Checksum: checksum,
	}
	if err := fileModel.Create(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("'%s' uploaded! checksum: %s", file.Filename, checksum),
	})
}

// Read - output file content
func (ctrl *FileController) Read(c *gin.Context) {
	var file model.File
	err := file.GetByChecksum(c.Query("s"))
	if err != nil {
		provider.DI().Log().Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.File(file.Path)
}
