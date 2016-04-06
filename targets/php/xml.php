<?

$content = file_get_contents('php://input');
echo $content, PHP_EOL;
echo "----------------------------------------------------------------------------", PHP_EOL;

function charXML($parser, $data)    {
  $data = trim($data);
  if($data == '' )
    return;

  echo $data, PHP_EOL;
}

$parser = xml_parser_create();
xml_set_character_data_handler($parser, 'charXML');
xml_parse($parser, $content);
xml_parser_free($parser);

?>