# gitlab-backup-tools

Backup all your gitlab repository with only 1 click !  
一键备份所有gitlab仓库

## Usage 如何使用

1. 准备 gitlab 用户的 accessToken， 并赋予所有 read 权限 `read_api, read_user, read_repository` - https://your.gitlab.domain/-/profile/personal_access_tokens
2. 修改 `config.go` 中的 gitlab 域名、gitlab accessToken
3. 安装依赖 `go mod tidy`
4. 运行 `go run main.go` 或 `go build . && ./gitlab-backup-tools`