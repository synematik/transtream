"use client";

import React from "react";

const Page: React.FC<{}> = ({}) => {
    async function onState(event: React.SyntheticEvent<HTMLVideoElement>): Promise<void> {
        const video: HTMLVideoElement = event.currentTarget || event.target;
        await fetch("http://127.0.0.1:8080/state", {
            method: "POST",
            body: JSON.stringify({
                state: !video.paused,
                time: video.currentTime,
            }),
            headers: {
                "Content-Type": "application/json",
            },
            mode: "no-cors",
        });
    }

    return (
        <video
            src='http://127.0.0.1:8080'
            className={'w-full h-full'}
            muted
            autoPlay
            controls
            controlsList="nodownload noremoteplayback noplaybackrate"
            disablePictureInPicture
            disableRemotePlayback
            // onPlay={onState}
            // onPause={onState}
        />
    );
}

export default Page;