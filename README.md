# gitlab-backup-tools

Backup all your gitlab repository with only 1 click !  
一键备份所有gitlab仓库

## Usage 如何使用

1. Prepare gitlab user's accessToken, and git read permission  `read_api, read_user, read_repository` - https://your.gitlab.domain/-/profile/personal_access_tokens  
> 准备 gitlab 用户的 accessToken， 并赋予所有 read 权限 `read_api, read_user, read_repository` - https://your.gitlab.domain/-/profile/personal_access_tokens
2. Modify the gitlab domain, gitlab accessToken in `config.go`  
> 修改 `config.go` 中的 gitlab 域名、gitlab accessToken
3. Run `go mod tidy` to install the dependencies  
> 安装依赖 `go mod tidy`
4. Run `go run main.go` or `go build . && ./gitlab-backup-tools` to start backup  
> 运行 `go run main.go` 或 `go build . && ./gitlab-backup-tools` 开始备份
5. The backup file will be placed at `downloaded_repos` folder， CI variables will be placed into each repository, named as `variables_group.json` and `variables_project.json`
> 备份文件将备份在 `downloaded_repos` 目录下, CI变量将被放在每个应用目录下，命名为 `variables_group.json` 和 `variables_project.json`