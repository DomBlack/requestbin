<?
// https://secure.php.net/manual/en/book.simplexml.php

if($_GET["url"]) {
  $content = file_get_contents($_GET["url"]);
} else {
  $content = file_get_contents('php://input');
}
echo $content, PHP_EOL;
echo "----------------------------------------------------------------------------", PHP_EOL;


if( $_GET["xxe"] == "true" ) {
  $doc = simplexml_load_string($content, null, LIBXML_NOENT);
} else {
  $doc = simplexml_load_string($content);
}

echo $doc, PHP_EOL;
?>