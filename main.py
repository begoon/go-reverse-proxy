from fastapi import FastAPI, Request
from fastapi.responses import PlainTextResponse

app = FastAPI()


@app.get("/{path:path}", response_class=PlainTextResponse)
async def root(request: Request, path: str):
    return f"I'm Python!\r\n[{path}]\r\n"
