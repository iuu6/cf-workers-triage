<?php

// 模拟 TOKEN，用于 Authorization 头
define('TOKEN', 'XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX');

// 模拟服务器地址
define('ADDRESS', 'https://XXXXXXXXX.XXXXX.XXX');

// 获取访问的URL
$currentUrl = $_SERVER['REQUEST_URI'];

// 从URL中提取文件路径
$path = parse_url($currentUrl, PHP_URL_PATH);

// 对路径进行rawurldecode解码
$path = rawurldecode($path);

// 构造请求头
$headers = [
    'Content-Type: application/json;charset=UTF-8',
    'Authorization: ' . TOKEN,
];

// 构造请求体
$data = [
    'path' => $path,
];

// 执行 POST 请求
$ch = curl_init(ADDRESS . '/api/fs/link');
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_CUSTOMREQUEST, 'POST');
curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($data));
curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);

// 获取响应
$response = curl_exec($ch);

// 处理响应
$httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);

// 如果HTTP Code是200，则进行重定向
if ($httpCode == 200) {
    $jsonResponse = json_decode($response, true);
    if (isset($jsonResponse['data']['url'])) {
        $redirectUrl = $jsonResponse['data']['url'];
        echo $redirectUrl;
        header("Location: $redirectUrl");
        exit();
    }
}

curl_close($ch);

// 输出响应信息
echo "HTTP Code: $httpCode\n";
echo "Response: $response\n";
?>
