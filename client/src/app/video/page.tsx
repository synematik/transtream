"use client";

import React from "react";

import {env} from "~/lib";

const Page: React.FC<{}> = ({}) => {
    async function onState(event: React.SyntheticEvent<HTMLVideoElement>): Promise<void> {
        const video: HTMLVideoElement = event.currentTarget || event.target;
        await fetch(`${env.API.BASE_URL}/state`, {
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
            src={env.API.BASE_URL}
            className={'w-full h-full'}
            muted
            autoPlay
            controls
            controlsList="nodownload noremoteplayback noplaybackrate"
            disablePictureInPicture
            disableRemotePlayback
            onPlay={onState}
            onPause={onState}
        />
    );
}

export default Page;