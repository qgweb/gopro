<?php

class Encrypt
{

    private $mcrypt_rijndael = MCRYPT_RIJNDAEL_256;
    private $mcrypt_mode = MCRYPT_MODE_CBC;
    private $mcrypt_rand = MCRYPT_RAND;
    protected static $_instance = NULL;


    public static function initialise()
    {
        if (NULL === self::$_instance) {
            self::$_instance = new self();
        }
        return self::$_instance;
    }

    public function encode($string, $key = '')
    {
        $key = md5($key);
        $enc = $this->mcrypt_encode($string, $key);
        return base64_encode($enc);
    }

    public function decode($string, $key = '')
    {
        $key = md5($key);

        if (preg_match('/[^a-zA-Z0-9\/\+=]/', $string)) {
            return FALSE;
        }

        $dec = base64_decode($string);

        if (($dec = $this->mcrypt_decode($dec, $key)) === FALSE) {
            return FALSE;
        }

        return $dec;
    }

    private function mcrypt_encode($data, $key)
    {
        $init_size = mcrypt_get_iv_size($this->mcrypt_rijndael, $this->mcrypt_mode);
        $init_vect = mcrypt_create_iv($init_size, $this->mcrypt_rand);
        return $this->add_cipher_noise($init_vect . mcrypt_encrypt($this->mcrypt_rijndael, $key, $data, $this->mcrypt_mode, $init_vect), $key);
    }

    private function mcrypt_decode($data, $key)
    {
        $data = $this->remove_cipher_noise($data, $key);
        $init_size = mcrypt_get_iv_size($this->mcrypt_rijndael, $this->mcrypt_mode);

        if ($init_size > strlen($data)) {
            return FALSE;
        }

        $init_vect = substr($data, 0, $init_size);
        $data = substr($data, $init_size);
        return rtrim(mcrypt_decrypt($this->mcrypt_rijndael, $key, $data, $this->mcrypt_mode, $init_vect), "\0");
    }

    private function add_cipher_noise($data, $key)
    {
        $keyhash = md5($key);
        $keylen = strlen($keyhash);
        $str = '';

        for ($i = 0, $j = 0, $len = strlen($data); $i < $len; ++$i, ++$j) {
            if ($j >= $keylen) {
                $j = 0;
            }

            $str .= chr((ord($data[$i]) + ord($keyhash[$j])) % 256);
        }

        return $str;
    }

    private function remove_cipher_noise($data, $key)
    {
        $keyhash = md5($key);
        $keylen = strlen($keyhash);
        $str = '';

        for ($i = 0, $j = 0, $len = strlen($data); $i < $len; ++$i, ++$j) {
            if ($j >= $keylen) {
                $j = 0;
            }

            $temp = ord($data[$i]) - ord($keyhash[$j]);

            if ($temp < 0) {
                $temp = $temp + 256;
            }

            $str .= chr($temp);
        }
        return $str;
    }

}

function fn_parse_url($str)
{
    $data = array();
    $parts = parse_url($str);
    if (isset($parts['query'])) {
        $tmp = explode('?', $str);
        $parameter = explode('&', end($tmp));
        foreach ($parameter as $val) {
            $index = strpos($val, "=");
            if ($index === false) {
                continue;
            }
            $sub1 = substr($val, 0, $index);
            $sub2 = substr($val, $index + 1);
            $sub2 = $sub2 === false ? "" : $sub2;
            $data[$sub1] = $sub2;
        }
    } else {
        $tmp = explode('?', $str);
        $parameter = explode('&', end($tmp));
        foreach ($parameter as $val) {
            $index = strpos($val, "=");
            if ($index === false) {
                continue;
            }
            $sub1 = substr($val, 0, $index);
            $sub2 = substr($val, $index + 1);
            $sub2 = $sub2 === false ? "" : $sub2;
            $data[$sub1] = $sub2;
        }
    }
    return $data;
}

if (count($argv) < 2) {
    echo "";
    exit;
}
$msg = base64_decode($argv[1]);
if (!$msg) {
    echo "";
    exit;
}

//$msg = '115.195.37.91	-	[06/Jul/2016:10:06:18 +0800]	"GET /t?u=IuIsPaJNZ%2BlSmYWQs7MWwAtc%2FkimWE4jIrzyYEDVvx3f5c4SGObmfYZkTesRivo4wxbM3qQUQe7MAa611F9EjRxF5a9YGfE%2F1ZP6cecYH9isKF6zIjC%2B1SVtF2NJRy0t1aFhhjS86tMArASwEE9tqQL7nR6byKENhmhHyuMbH3IYXLCG1XpD3I8wE2BZndoH%2BQEfaem%2FTYLILuFafzQZyNWBgAhBhcBqTdefprwy3MrqwmOFujjIOFo302bzUor6UvpPuozkSesSUDNv1XGBxzB46kCZBUqEloTTaRlIVvjRRd%2BoB4qA70aAevycyGx%2BNBu44nC7LhxXSM9rRp6r5vKFSlcwFrJHgPH%2BD%2FvxUqgf%2FeOG3079M4OEJYrF8fg7OBGlNGA161Y8WjRl2nDBYLO59Bkh%2FnkvlZz3FDVHjC8JeUWPZOixth5zngJgmfAf8a4dPtmtd1KhWNONpRf6pS9aCvB5pcrnPyesgm3cS25obo0EesowziRNYxHQheaI&ktt=cnb_1467769281861_flag HTTP/1.1"	302	5	"http://cpro.9xu.com/tjs?tu=http%3A%2F%2Fcpro.9xu.com%2Ft%3Fu%3DIuIsPaJNZ%252BlSmYWQs7MWwAtc%252FkimWE4jIrzyYEDVvx3f5c4SGObmfYZkTesRivo4wxbM3qQUQe7MAa611F9EjRxF5a9YGfE%252F1ZP6cecYH9isKF6zIjC%252B1SVtF2NJRy0t1aFhhjS86tMArASwEE9tqQL7nR6byKENhmhHyuMbH3IYXLCG1XpD3I8wE2BZndoH%252BQEfaem%252FTYLILuFafzQZyNWBgAhBhcBqTdefprwy3MrqwmOFujjIOFo302bzUor6UvpPuozkSesSUDNv1XGBxzB46kCZBUqEloTTaRlIVvjRRd%252BoB4qA70aAevycyGx%252BNBu44nC7LhxXSM9rRp6r5vKFSlcwFrJHgPH%252BD%252FvxUqgf%252FeOG3079M4OEJYrF8fg7OBGlNGA161Y8WjRl2nDBYLO59Bkh%252FnkvlZz3FDVHjC8JeUWPZOixth5zngJgmfAf8a4dPtmtd1KhWNONpRf6pS9aCvB5pcrnPyesgm3cS25obo0EesowziRNYxHQheaI&t=1467770729"	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/49.0.2623.110 Safari/537.36"	-	CNZZDATA1256979552=1778728267-1460615251-%7C1460615251; CNZZDATA1258855680=1103945022-1461821395-%7C1463550630; Hm_lvt_1941919deb79f5279264250ca58002d5=1462242611,1463398694,1463548938,1464068026; Hm_lvt_c81476f1761060d378609277ea07b570=1462242612,1463398695,1463548938,1464068026; CNZZDATA1257915830=317006853-1461232279-%7C1464247532; CNZZDATA1259003728=1151216506-1462518438-%7C1464843678; CNZZDATA1257360444=495175581-1460613525-%7C1466744435; _ga=GA1.2.1809719460.1466748101; OUTFOX_SEARCH_USER_ID_NCOO=1685763349.064891; cnb_1467769281861_flag=7dec2e2169de9673dfe92efd0e3dd8a8; CNZZDATA1254772867=1198470891-1461128993-%7C1467769552; dt_uid=577229c16adc0fec538b4567; dt_cox=aaaaaaaaaaa; dt_uptime=1467770728; dt_area=nil; CNZZDATA5948225=cnzz_eid%3D372871705-1460612110-http%253A%252F%252Fcpro.9xu.com%252F%26ntime%3D1467766163; CNZZDATA1258321956=1384707831-1460610356-http%253A%252F%252Fcpro.9xu.com%252F%7C1467767048';
$info = explode("\t", $msg);
if (count($info) < 3) {
    echo "";
    exit;
}

$u = explode(" ", $info[3])[1];
$pu = fn_parse_url($u);
if (!isset($pu['u'])) {
    echo "";
    exit;
}
$msg = Encrypt::initialise()->decode(urldecode($pu['u']), "&^*84558&#1534#$!");
if (!$msg) {
    echo "";
    exit;
}

$pmsg = fn_parse_url($msg);
echo json_encode($pmsg);
exit;
$cox = "";

if (isset($pmsg["cox"]) && !empty($pmsg["cox"])) {
    $cox = $pmsg["cox"];
}
if (isset($pmsg["js_cox"]) && !empty($pmsg["js_cox"])) {
    $cox = $pmsg["js_cox"];
}
if (isset($pmsg["js2_cox"]) && !empty($pmsg["js2_cox"])) {
    $cox = $pmsg["js2_cox"];
}
if (isset($pmsg["sh_cox"]) && !empty($pmsg["sh_cox"])) {
    $cox = $pmsg["sh_cox"];
}

echo json_encode(array(
    'ua' => substr($info[7], 1, strlen($info[7]) - 2),
    'ad' => urldecode($cox),
    'hd' => $pmsg["hd"],
    'pd' => $pmsg["pd"],
    'lftu' => urldecode($pmsg['lftu']),
    'ltu' => urldecode($pmsg['ltu']),
));