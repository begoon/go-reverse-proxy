const { createServer } = require("http");

createServer((req, res) => {
    res.setHeader("content-type", "text/plain");
    res.end(`I'm Node!\r\n[${req.url}]\r\n`);
}).listen(process.env.PORT || 9000);
