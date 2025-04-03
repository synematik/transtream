ffmpeg -stream_loop -1 -re -i video.mp4 -an -c:v libx264 -f rtp rtp://127.0.0.1:5004
ffmpeg -stream_loop -1 -re -i video.mp4 -vn -c:a libopus -f rtp rtp://127.0.0.1:5006