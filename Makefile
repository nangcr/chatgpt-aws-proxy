# 设置 Lambda 函数的名称
FUNCTION_NAME := 你的Lambda函数名称

# 设置 Lambda 函数所在的 AWS 区域
AWS_REGION := us-east-1

# 设置用于存储 Lambda 部署包的 AWS S3 存储桶的名称
S3_BUCKET := 你的S3存储桶名称

# 设置部署包的名称
DEPLOYMENT_PACKAGE := deployment-package.zip

# 构建部署包
build:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o main
	zip -r $(DEPLOYMENT_PACKAGE) main

# 部署 Lambda 函数
deploy: build
	aws lambda create-function \
		--region $(AWS_REGION) \
		--function-name $(FUNCTION_NAME) \
		--handler main \
		--runtime go1.x \
		--role your-aws-lambda-role-arn \
		--zip-file fileb://$(DEPLOYMENT_PACKAGE)

# 更新 Lambda 函数
update: build
	aws lambda update-function-code \
		--region $(AWS_REGION) \
		--function-name $(FUNCTION_NAME) \
		--zip-file fileb://$(DEPLOYMENT_PACKAGE)

# 删除 Lambda 函数
delete:
	aws lambda delete-function \
		--region $(AWS_REGION) \
		--function-name $(FUNCTION_NAME)

# 将部署包上传到 S3
upload:
	aws s3 cp $(DEPLOYMENT_PACKAGE) s3://$(S3_BUCKET)/$(DEPLOYMENT_PACKAGE)

# 从本地文件系统中删除部署包
clean:
	rm -f main $(DEPLOYMENT_PACKAGE)
