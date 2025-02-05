const { NodeMediaServer } = require('node-media-server');

const config = {
    logType: 3,
    rtmp: {
        port: 1935,
        chunk_size: 60000,
        gop_cache: true,
        ping: 60,
        ping_timeout: 30
    },
    http: {
        port: 8000,
        allow_origin: '*'
    },
    relay: {
        ffmpeg: '/usr/local/bin/ffmpeg',
        tasks: [
            {
                app: 'cctv',
                mode: 'static',
                edge: 'rtsp://localhost:8554/mystream',
                name: 'uterum',
                rtsp_transport : 'tcp'
            }
        ]
    }
};

const nms = new NodeMediaServer(config);
nms.run();
