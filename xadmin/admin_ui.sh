# 删除并复制文件
rm -rf ui/* && cp -rf ./adminui/dist/* ./ui/

# 替换 href="/ 为 href="
sed -i '' 's|href="/|href="|g' ./ui/index.html
sed -i '' 's|src="/|src="|g' ./ui/index.html

# 替换 <title></title> 为 <script>...</script>
sed -i '' 's|<title></title>|<script> window.BASE_URL = \"/\" + location.pathname.split(\"/\").filter(e => e).concat([\"api\"]).join(\"\")</script>|g' ./ui/index.html
