
Certificates setup [optional]:
```shell
$ brew install mkcert

$ mkcert -install
Created a new local CA 💥

The local CA is now installed in the system trust store! ⚡️
Warning: "certutil" is not available, so the CA can't be automatically installed in Firefox! ⚠️
Install "certutil" with "brew install nss" and re-run "mkcert -install" 👈

$ brew install nss

$ mkcert -install 
The local CA is already installed in the system trust store! 👍
The local CA is now installed in the Firefox trust store (requires browser restart)! 🦊

$ mkcert example.com "*.example.com" example.test localhost 127.0.0.1 ::1

Created a new certificate valid for the following names 📜
 - "example.com"
 - "*.example.com"
 - "example.test"
 - "localhost"
 - "127.0.0.1"
 - "::1"

Reminder: X.509 wildcards only go one level deep, so this won't match a.b.example.com ℹ️

The certificate is at "./example.com+5.pem" and the key at "./example.com+5-key.pem" ✅

It will expire on <some date> 🗓

 # Confirm
$ tree -L 1 | grep example
├── example.com+5-key.pem
├── example.com+5.pem


```
