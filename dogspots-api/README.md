# DogSpots API

La API carga automáticamente datos de ejemplo cuando `APP_ENV=development`.
La semilla idempotente incluye cuatro parques de Madrid; para evitar duplicados,
se pueden conservar al reiniciar la API. En producción, configura
`APP_ENV=production` para no cargar esos datos.

Los nombres y coordenadas de referencia proceden de OpenStreetMap (© OpenStreetMap
contributors, ODbL). Los puntos representan ubicaciones aproximadas dentro de cada
parque, no accesos ni zonas caninas oficiales. Comprueba las normas locales sobre
perros y uso de correa.

Para el Nginx que se ejecuta en el host, añade este bloque dentro del `server`
HTTPS para hacer proxy de las fotos a la API, publicada solo en loopback. La API
establece `Cache-Control: public, max-age=31536000, immutable` y Nginx conserva
esa cabecera en la respuesta:

```nginx
location ^~ /media/ {
	proxy_pass http://127.0.0.1:8080;
	proxy_set_header Host $host;
	proxy_set_header X-Real-IP $remote_addr;
	proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
	proxy_set_header X-Forwarded-Proto $scheme;
}
```