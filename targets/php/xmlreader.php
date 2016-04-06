<?

$content = file_get_contents('php://input');
echo $content, PHP_EOL;
echo "----------------------------------------------------------------------------", PHP_EOL;

$xml = new XMLReader();
$xml->open('php://input');

while ($xml->read()) {
  if($xml->nodeType == XMLReader::TEXT) {
    echo $xml->value, PHP_EOL;
  }
}

?>
