<?php

function sendLog(string $text, bool $encrypt = false): string
{
    $result = "";

    $socket = socket_create(AF_INET, SOCK_STREAM, SOL_TCP);

    if (is_resource($socket)) {
        throw new \Exception("Create socket is failed", 500);
    }

    socket_connect($socket, '127.0.0.1', '9001');

    // prefix '1': cipher text; prefix '0': plain text
    if ($encrypt) {
        $text = '1' . $text;
    } else {
        $text = '0' . $text;
    }
    socket_write($socket, $text);

    $result = socket_read($socket, 1024);

    socket_close($socket);

    return $result;
}

while (true) {
    $num = mt_rand(0, 100);

    $response = sendLog(json_encode([
        'name' => 'leo',
        'num' => $num,
    ]));

    echo sprintf("num: %s, response: %s", $num, $response), PHP_EOL;

    sleep(1);
}
