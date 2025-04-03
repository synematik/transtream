ffmpeg -re -i video.mp4 -c:v libx264 -preset ultrafast -tune zerolatency -f rtp "rtp://127.0.0.1:5004?pkt_size=1200"
ffmpeg -i "video.mp4" -f rtp "rtp://127.0.0.1:5005?pkt_size=1200"

ffmpeg -stream_loop -1 -re -i "video.mp4" -c:v h264_nvenc -preset p7 -profile:v baseline -level 4.0 -pix_fmt yuv420p -forced-idr 1 -g 1 -sc_threshold 0 -c:a aac -ar 44100 -b:a 128k -ac 2 -f rtp_mpegts -flush_packets 0 -fflags +genpts "rtp://239.255.12.42:5000?localrtpport=5005&ttl=1&pkt_size=1316"

#! works
ffmpeg -stream_loop -1 -re -i "video.mp4" -c:v libx264 -preset ultrafast -tune zerolatency -f rtp_mpegts "rtp://127.0.0.1:5004?pkt_size=1200"


ffmpeg -stream_loop -1 -re -i video.mp4 -c:v libx264 -profile:v baseline -level 3.0 -pix_fmt yuv420p -an -payload_type 96 -f rtp rtp://127.0.0.1:5000 -c:a libopus -vn -payload_type 111 -f rtp rtp://127.0.0.1:5002 > stream.sdp