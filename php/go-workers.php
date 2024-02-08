<?php

// 读取data.json文件内容
$jsonData = file_get_contents('data.json');

// 将JSON数据解析为关联数组
$data = json_decode($jsonData, true);

if ($data === null) {
    die('Failed to parse JSON data');
}

// 找到requests最少的pattern
$minRequestsPattern = '';
$minRequests = PHP_INT_MAX;

foreach ($data as $item) {
    if ($item['requests'] < $minRequests) {
        $minRequests = $item['requests'];
        $minRequestsPattern = $item['pattern'];
    }
}

// 判断requests是否大于等于10万
if ($minRequests >= 100000) {
    die('Data size is too large');
}

// 获取当前请求的路径
$currentPath = parse_url($_SERVER['REQUEST_URI'], PHP_URL_PATH);
$queryString = parse_url($_SERVER['REQUEST_URI'], PHP_URL_QUERY);

// 构建重定向URL
$redirectUrl = $minRequestsPattern . $currentPath;

// 如果有查询参数，添加到重定向URL
if ($queryString !== null) {
    $redirectUrl .= '?' . $queryString;
}

// 获取协议和主机部分
$originalUrl = parse_url($_SERVER['REQUEST_URI']);
$protocol = isset($originalUrl['scheme']) ? $originalUrl['scheme'] : 'https';
$host = isset($originalUrl['host']) ? $originalUrl['host'] : '';

// 发送302重定向
header("Location: $protocol://$host$redirectUrl", true, 302);
exit;
