# Server:
### [OKK] HTTP *singleclient
ffmpeg -i "https://synema.cxllmerichie.com/proxy/6e0589342c84c1e468c6442bad7cfbf4:2025020707:R01lcjFaQkF1QXFCeHBCY20weGU0WVh1am5HVzVZT0swcElWN3k2M1hja2hPVURhdlFLd2xobHluODRkd2hydFFtS2lSRGZTTC9RQVdRRjBzNzNtanc9PQ==/2/4/8/7/3/5/brh53.mp4:hls:manifest.m3u8" -listen 1 -f mp4 -movflags frag_keyframe+empty_moov http://localhost:8080

### [ERR] UDP
ffmpeg -re -i https://synema.cxllmerichie.com/proxy/6e0589342c84c1e468c6442bad7cfbf4:2025020707:R01lcjFaQkF1QXFCeHBCY20weGU0WVh1am5HVzVZT0swcElWN3k2M1hja2hPVURhdlFLd2xobHluODRkd2hydFFtS2lSRGZTTC9RQVdRRjBzNzNtanc9PQ==/2/4/8/7/3/5/brh53.mp4:hls:manifest.m3u8 -vf scale=1280:720 -vcodec libx264 -pix_fmt yuv420p -tune zerolatency -preset ultrafast -f mpegts udp://236.0.0.1:2000