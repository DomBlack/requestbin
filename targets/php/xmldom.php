<?
// https://secure.php.net/manual/en/domdocument.loadxml.php

if($_GET["url"]) {
  $content = file_get_contents($_GET["url"]);
} else {
  $content = file_get_contents('php://input');
}
echo $content, PHP_EOL;
echo "----------------------------------------------------------------------------", PHP_EOL;


$dom = new DOMDocument();
if( $_GET["xxe"] == "true" ) {
  $dom->loadXML($content, LIBXML_NOENT);
} else {
  $dom->loadXML($content);
}

echo $dom->saveXML();

?>