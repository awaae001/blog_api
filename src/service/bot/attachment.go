package bot

import (
	"blog_api/src/service"
	"blog_api/src/service/oss"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// storeAttachment tries OSS first and rewinds the stream before local fallback.
func storeAttachment(svc oss.OSSService, name, mimeType string, file io.ReadSeeker, size int64) (string, int, error) {
	subPath := filepath.Join("moments", time.Now().Format("060102"))
	if svc != nil {
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			return "", 0, err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		url, _, err := svc.Upload(ctx, oss.UploadInput{
			Name:        filepath.Join(subPath, name),
			ContentType: mimeType,
			Size:        size,
			Body:        file,
		})
		cancel()
		if err == nil {
			return url, 0, nil
		}
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", 0, err
	}
	_, url, err := service.SaveResourceReader(name, file, subPath, false)
	return url, 1, err
}

const maxBotAttachmentBytes = int64(64 << 20)

var attachmentHTTPClient = &http.Client{Timeout: 2 * time.Minute}

func downloadToTemp(ctx context.Context, rawURL string) (*os.File, int64, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("create attachment request: %w", err)
	}
	resp, err := attachmentHTTPClient.Do(req)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("download attachment: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, 0, nil, fmt.Errorf("download attachment: status %s", resp.Status)
	}
	if resp.ContentLength > maxBotAttachmentBytes {
		return nil, 0, nil, fmt.Errorf("attachment exceeds %d bytes", maxBotAttachmentBytes)
	}

	file, err := os.CreateTemp("", "blog-api-media-*")
	if err != nil {
		return nil, 0, nil, fmt.Errorf("create attachment temp file: %w", err)
	}
	keep := false
	defer func() {
		if !keep {
			file.Close()
			os.Remove(file.Name())
		}
	}()

	size, err := io.Copy(file, io.LimitReader(resp.Body, maxBotAttachmentBytes+1))
	if err != nil {
		return nil, 0, nil, fmt.Errorf("write attachment temp file: %w", err)
	}
	if size > maxBotAttachmentBytes {
		return nil, 0, nil, fmt.Errorf("attachment exceeds %d bytes", maxBotAttachmentBytes)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, 0, nil, fmt.Errorf("rewind attachment: %w", err)
	}

	sample := make([]byte, min(size, 512))
	if _, err := io.ReadFull(file, sample); err != nil {
		return nil, 0, nil, fmt.Errorf("read attachment sample: %w", err)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, 0, nil, fmt.Errorf("rewind attachment: %w", err)
	}
	keep = true
	return file, size, sample, nil
}
