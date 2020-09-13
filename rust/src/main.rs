#![warn(clippy::all, missing_debug_implementations, rust_2018_idioms)]

use futures::future::select_all;
use futures::FutureExt;
use serde::Deserialize;
use std::collections::HashSet;
use std::sync::{Arc, Mutex};
use tokio::macros::support::Future;

type Error = Box<dyn std::error::Error>;
type Result<R, E = Error> = std::result::Result<R, E>;

const URL: &str = "https://graph.modulitos.com";

struct RewardIncrementer {
    mutex: Mutex<u32>,
}

impl RewardIncrementer {
    fn increment(&self, reward: u32) {
        let mut lock = self.mutex.lock().unwrap();
        *lock += reward;
    }
}

type Node = char;

#[derive(Default)]
struct NodeTracker {
    // These are nodes that have already been fetched and processed.
    visited: Arc<Mutex<HashSet<Node>>>,
    // These are the next nodes that we want to fetch:
    next: Arc<Mutex<HashSet<Node>>>,
}

impl NodeTracker {
    fn add_node(&mut self, node: Node) {
        let visited = self.visited.lock().unwrap();
        let mut next = self.next.lock().unwrap();
        if !visited.contains(&node) && !next.contains(&node) {
            next.insert(node);
        }
    }

    fn process_node(&self, node: Node) -> impl Future {
        async move {
            let res = fetch_node(node)
                .await
                .expect("HTTP request to fetch and deserialize the node failed!");
        }
    }

    fn transition_next_to_futures(&mut self) -> Vec<impl Future> {
        let next = std::mem::replace(&mut self.next, Arc::new(Mutex::new(HashSet::new())));

        // TODO: try using into_inner() to move the underlying data out of the Arc<Mutex>
        let next = next.lock().unwrap();
        next.iter()
            .map(|&node| self.process_node(node).boxed())
            .collect()
    }
}

#[derive(Deserialize, Debug)]
struct Response {
    children: Vec<Node>,
    reward: u32,
}
#[tokio::main]
async fn main() -> Result<()> {
    let totals = RewardIncrementer {
        mutex: Mutex::new(0),
    };
    let mut tracker = NodeTracker::default();
    tracker.add_node('a');

    let mut pending_futures = Vec::new();

    loop {
        let mut next_futures = tracker.transition_next_to_futures();
        pending_futures.append(&mut next_futures);
        if !pending_futures.is_empty() {
            let (item_resolved, _index, pending) = select_all(pending_futures).await;
            pending_futures = pending;
        } else {
            break;
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
    assert_eq!(*total, 0);
    Ok(())
}

async fn fetch_node(
    node: Node,
) -> Result<Response> {
    let resp = reqwest::get(&format!("{}/node/{}", URL, node))
        .await?
        .json::<Response>()
        .await?;
    println!("resp: {:?}", resp);
    Ok(resp)
}
