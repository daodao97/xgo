# 生成私钥
openssl genrsa -out private.pem 2048

# 从私钥中提取公钥
openssl rsa -in private.pem -pubout -out public.pem

# 查看私钥内容
cat private.pem

# 查看公钥内容
cat public.pem