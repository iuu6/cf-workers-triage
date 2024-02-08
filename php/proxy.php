<?php

// 读取data.json文件内容
$jsonData = file_get_contents('data.json');

// 将JSON数据解析为关联数组
$data = json_decode($jsonData, true);

if ($data === null) {
    die('Failed to parse JSON data');
}

// 获取当前请求的路径
$currentPath = parse_url($_SERVER['REQUEST_URI'], PHP_URL_PATH);
$queryString = parse_url($_SERVER['REQUEST_URI'], PHP_URL_QUERY);

// 检查文件名是否包含指定后缀
$allowedExtensions = ['txt', 'htm', 'xml', 'java', 'properties', 'sql', 'js', 'md', 'json', 'conf', 'ini', 'vue', 'php', 'py', 'bat', 'gitignore', 'yml', 'go', 'sh', 'c', 'cpp', 'h', 'hpp', 'tsx', 'vtt', 'srt', 'ass', 'rs', 'lrc'];
$fileName = pathinfo($currentPath, PATHINFO_FILENAME);
$fileExtension = pathinfo($currentPath, PATHINFO_EXTENSION);

if (in_array($fileExtension, $allowedExtensions) || strpos($fileName, '.') !== false) {
    // 如果文件名包含指定后缀或者包含点号，则直接处理请求
    handleRequest();
} else {
    // 否则，进行302转发
    perform302Redirect();
}

function handleRequest() {
    global $data, $currentPath, $queryString;

    // 找到requests最少的pattern
    $minRequestsPattern = '';
    $minRequests = PHP_INT_MAX;

    foreach ($data as $item) {
        if ($item['requests'] < $minRequests) {
            $minRequests = $item['requests'];
            $minRequestsPattern = $item['pattern'];
        }
    }

    // 构建目标URL
    $targetUrl = $minRequestsPattern . $currentPath;

    // 如果有查询参数，添加到目标URL
    if ($queryString !== null) {
        $targetUrl .= '?' . $queryString;
    }

    // 获取协议和主机部分
    $originalUrl = parse_url($_SERVER['REQUEST_URI']);
    $protocol = isset($originalUrl['scheme']) ? $originalUrl['scheme'] : 'https';
    $host = isset($originalUrl['host']) ? $originalUrl['host'] : '';

    // 构建cURL句柄
    $ch = curl_init();

    // 设置cURL参数
    curl_setopt($ch, CURLOPT_URL, "$protocol://$host$targetUrl");
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
    // 在此可以设置其他cURL选项，如需要的话

    // 执行cURL请求并获取结果
    $result = curl_exec($ch);

    // 检查是否有错误发生
    if (curl_errno($ch)) {
        die('cURL error: ' . curl_error($ch));
    }

    // 关闭cURL句柄
    curl_close($ch);

    // 设置正确的HTTP头部，告诉浏览器这是一个可下载的文件
    header('Content-Description: File Transfer');
    header('Content-Type: application/octet-stream');
    header('Content-Disposition: attachment; filename="' . basename($currentPath) . '"');
    header('Expires: 0');
    header('Cache-Control: must-revalidate');
    header('Pragma: public');
    header('Content-Length: ' . strlen($result));

    // 直接输出cURL请求的结果
    echo $result;
    exit;
}



function perform302Redirect() {
    global $data;

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
