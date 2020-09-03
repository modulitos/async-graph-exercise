#![warn(clippy::all, missing_debug_implementations, rust_2018_idioms)]

use serde::{Deserialize, Serialize};
use std::sync::Mutex;
use std::collections::VecDeque;
use tokio::macros::support::Future;
use tokio::sync::mpsc;
use std::borrow::Borrow;
use std::ops::Deref;

type Error = Box<dyn std::error::Error>;
type Result<R, E = Error> = std::result::Result<R, E>;

const URL: &str = "https://graph.modulitos.com";

struct RewardIncrementer {
    mutex: Mutex<u32>
}

impl RewardIncrementer {
    fn increment(&self, reward: u32) {
        let mut lock = self.mutex.lock().unwrap();
        *lock += reward;
    }
}

#[derive(Deserialize, Debug)]
struct Response {
    children: Vec<char>,
    reward: u32,
}
#[tokio::main]
async fn main() -> Result<()>{

    let totals = RewardIncrementer {
        mutex: Mutex::new(0)
    };

    let mut futures = VecDeque::new();
    futures.push_back(fetch_node('a'));

    loop {

        // TODO: This is only awaiting one future at a time. To fix this, we'll need to update
        // `futures` to store all of the futures, and wait until it is empty. If not empty, then
        // select the first future that is ready.

        if let Some(future) = futures.pop_front() {
            let resp = future.await?;
            totals.increment(resp.reward);
            let mut new_futures = resp.children.into_iter().map(

                // TODO: Consider doing tokio::spawn(async {...}) here to fetch the node, then
                // update shared state. That will allow us to run the futures concurrently, instead
                // of having to queue them up.

                |c| fetch_node(c)
            ).collect::<VecDeque<_>>();
            futures.append(&mut new_futures);
        } else {
            break
        }
    }

    // possible approach, using channels:

    // let (tx1, mut rx1) = mpsc::channel(128);


    // tokio::spawn(async move {
    //     // Send values on `tx1`
    //     tx1.clone().send("node value response").await.unwrap();
    // });
    //
    // loop {
    //     // Run until the channel is closed. (when is the channel closed?)
    //     let msg = tokio::select! {
    //         Some(msg) = rx1.recv() => msg,
    //         else => { break }
    //     };
    //
    //     println!("Got {}", msg);
    // }

    let total = totals.mutex.lock().unwrap();
    println!("total: {}", total);
    assert_eq!(*total, 4250);
    Ok(())
}

async fn fetch_node(node: char) -> Result<Response> {
    let resp = reqwest::get(&format!("{}/node/{}", URL, node))
        .await?
        .json::<Response>()
        .await?;
    println!("resp: {:?}", resp);
    Ok(resp)
}
