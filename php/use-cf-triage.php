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

// 判断是否为中国或者requests是否大于等于9万
$isChina = isset($_SERVER['HTTP_CF_IPCOUNTRY']) && $_SERVER['HTTP_CF_IPCOUNTRY'] === 'CN';
if (!$isChina || $minRequests >= 90000) {
    // 执行 CF-IPCountry 不是中国 或者 requests 大于等于 9 万的代码块
    // ...

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
} else {
    // CF-IPCountry 是中国 且 requests 小于 9 万的代码块
    // ...

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
}
?>
