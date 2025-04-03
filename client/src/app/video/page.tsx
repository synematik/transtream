"use client";

import React from "react";

import {env} from "~/lib";

const Page: React.FC<{}> = ({}) => {
    return (
        <video
            src={`http://127.0.0.1:8079/`}
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