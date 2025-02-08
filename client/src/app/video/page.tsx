"use client";

import React from "react";

import {env} from "~/lib";

const Page: React.FC<{}> = ({}) => {
    return (
        <video
            src={`${env.API.BASE_URL}/stream`}
            className={'w-full h-full'}
            muted
            autoPlay
            controls
            controlsList="nodownload noremoteplayback noplaybackrate"
            disablePictureInPicture
            disableRemotePlayback
        />
    );
}

export default Page;