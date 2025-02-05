mod api;
mod config;
mod context;
mod ffmpeg;
mod model;
mod scan;
mod sqlite;
mod tmdb;

use crate::api::MakeApiSvc;
use crate::config::Config;
use crate::context::Context;
use crate::model::{Table};
use hyper::Server;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let ctx = Context::from_config(Config::new());
    let config = ctx.cfg();

    let addr = ([127, 0, 0, 1], 3000).into();

    let server = Server::bind(&addr).serve(MakeApiSvc::new(config, sqlite));
    println!("Listening on http://{}", addr);

    server.await?;
    Ok(())
}
