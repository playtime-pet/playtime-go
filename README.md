before running this project, you need to set the following environment variables:
WECHAT_APPID
WECHAT_SECRET
WECHAT_MINI_MAP_API
MONGO_USER
MONGO_PASS
COS_SECRET_ID
COS_SECRET_KEY
COS_BUCKET_URL

run command to build the file

```shell
GOOS=linux GOARCH=amd64 go build -o playtime-go

ssh ubuntu@playtime "rm -f ~/playtime/playtime-go"
scp playtime-go ubuntu@playtime:~/playtime/

ps aux | grep playtime-go

nohup ~/playtime/playtime-go > ~/playtime/out.log 2>&1 &
```
