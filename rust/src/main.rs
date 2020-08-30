#![warn(clippy::all, missing_debug_implementations, rust_2018_idioms)]

use serde::{Deserialize, Serialize};

const URL: &str = "https://graph.modulitos.com";

#[derive(Deserialize, Debug)]
struct Response {
    children: Vec<String>,
    reward: u16,
}

#[tokio::main]
async fn main() {
    println!("starting");
    fetch_node('a').await.unwrap();
    println!("done!");
}

async fn fetch_node(node: char) -> Result<(), Box<dyn std::error::Error>> {
    let resp = reqwest::get(&format!("{}/node/{}", URL, node))
        .await?
        .json::<Response>()
        .await?;
    println!("resp: {:?}", resp);
    Ok(())
}
