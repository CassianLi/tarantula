# tarantula（狼蛛）

`Trarantula` 通过监听消息队列中的截图请求，从而使用`selenium` 库对`Ebay` 商品页进行截图。将截图上传到`Ali OSS` 存储，并将截图操作的执行情况已消息的方式发送给 `Rabbit MQ`指定队列。

## Documentation

### Requirements

- [github.com/streadway/amqp](https://github.com/streadway/amqp)
- [github.com/tebeka/selenium](https://github.com/tebeka/selenium)
- [github.com/aliyun/aliyun-oss-go-sdk/oss](https://github.com/baiyubin/aliyun-sts-go-sdk)
- [gopkg.in/ini.v1](https://gopkg.in/ini.v1)

### Install

#### Build

```shell
git clone 

cd tarantula/

go build
```

#### Config file

程序运行前需要配置必要参数，并保存至以`.ini` 后缀的配置文件中，默认读取`./conf.ini` 文件中的参数。你也可以自定配置文件名。配置文件内容如下：

```ini
# possible values : production, development
AppMode = development
[Queue]
# Queue is used to Listen
Consume = your-consume-queue-name
# Queue is used to Publish screenshot result
Publish = your-publish-queue-name

[Rabbit]
Url = amqp://your-mq-user:your-mq-password@127.0.0.1:5672
Exchange = exchange-name

[Selenium]
# browser driver path
DriverPath = /opt/homebrew/bin/geckodriver
# browser driver start port,chrome:8080, firefox:4444
Port = 4444

# your ali OSS
[OSS]
Endpoint = your-ali-oss-endpoint
AccessID = your-ali-oss-accessID
AccessKey = your-ali-oss-accessKey
BucketName = your-ali-oss-bucketName

```

#### Run

将编译后的二进制文件`tarantula -c conf.ini` 直接运行即可开启命令。但仍建议使用`systemctl` 进行服务管理

### Systemctl Install

#### 1. 添加`systemctl` 服务配置文件

编辑文件`/etc/systemd/system/tarantula.service` ,在其中添加如下内容

```shell
[Unit]
Description = Tarantula service, used to make ebay web screenshots
After = network.target syslog.target
Wants = network.target

[Service]
Type = simple
ExecStart = /your-path/tarantula -c /your-path/tarantula.ini

[Install]
WantedBy = multi-user.target
```