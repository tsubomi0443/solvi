package converter

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"path/filepath"
	outputmodel "solvi/internal/application/model/output_model"
	"solvi/internal/domain/entity"
	"strings"
)

func UserEntityToOutput(u *entity.User) outputmodel.UserOutput {
	icon := ""
	base64Text := ""
	if u.Icon != nil {
		icon = *u.Icon
		ext := filepath.Ext(*u.Icon)
		base64Prefix := getExtPrefix(ext)
		base64Icon := base64.StdEncoding.EncodeToString(u.IconBlob)
		base64Text = fmt.Sprintf("%s,%s", base64Prefix, base64Icon)
	}

	return outputmodel.UserOutput{
		UUID:           u.UUID.String(),
		Name:           u.Name,
		Email:          u.Email,
		DepartmentName: u.DepartmentName,
		Icon:           icon,
		IconBase64:     template.URL(base64Text),
		IsSupporter:    u.IsSupporter,
		IsAdmin:        u.IsAdmin,
	}
}

func getExtPrefix(ext string) string {
	cleanExt := strings.ToLower(strings.TrimPrefix(ext, "."))
	switch cleanExt {
	case "jpg", "jpeg":
		return "data:image/jpeg;base64"
	case "png":
		return "data:image/png;base64"
	case "webp":
		return "data:image/webp;base64"
	case "svg":
		return "data:image/svg+xml;base64"
	default:
		// 未知の形式は一般的なバイナリストリームとして扱うか、jpeg等にフォールバック
		return "data:application/octet-stream;base64"
	}
}
