#![warn(clippy::all, missing_debug_implementations, rust_2018_idioms)]

use futures::future::{select_all, BoxFuture};
use futures::FutureExt;
use serde::Deserialize;
use std::collections::HashSet;
use std::sync::{Arc, Mutex};
use tokio::macros::support::Future;

type Error = Box<dyn std::error::Error>;
type Result<R, E = Error> = std::result::Result<R, E>;

const URL: &str = "https://graph.modulitos.com";

type Reward = u32;

#[derive(Default)]
struct RewardIncrementer {
    mutex: Mutex<Reward>,
}

impl RewardIncrementer {
    fn increment(&self, reward: Reward) {
        let mut lock = self.mutex.lock().unwrap();
        *lock += reward;
    }

    fn value(&self) -> Reward {
        *self.mutex.lock().unwrap()
    }
}

type Node = char;

#[derive(Default)]
/// Handles the nodes
struct NodeTracker {
    // Nodes that have already been fetched and processed.
    visited: Arc<Mutex<HashSet<Node>>>,
    // The next nodes that we want to fetch:
    next: Arc<Mutex<HashSet<Node>>>,
    // Used for incrementing the rewards associated with each node:
    rewards: Arc<RewardIncrementer>,
}

impl NodeTracker {
    fn add_node(&mut self, node: Node) {
        let visited = self.visited.lock().unwrap();
        let mut next = self.next.lock().unwrap();
        if !visited.contains(&node) && !next.contains(&node) {
            next.insert(node);
        }
    }

    fn process_node(
        &self,
        node: Node,
    ) -> impl Future {
        let visited = self.visited.clone();
        let next = self.next.clone();
        let incrementer = self.rewards.clone();
        async move {
            let resp = fetch_node(node)
                .await
                .expect("HTTP request to fetch and deserialize the node failed!");

            println!("resp from '{}': {:?}", node, resp);

            incrementer.increment(resp.reward);

            // Add the children to the "next" set, but only if they are not already within the
            // "visited" set.
            let mut visited_slot = visited.lock().unwrap();
            visited_slot.insert(node);

            let mut nodes_to_add = resp
                .children
                .into_iter()
                .filter(|child| !visited_slot.contains(child));

            next.lock().unwrap().extend(&mut nodes_to_add);
            println!("next: {:?}", next);
            println!();
        }
    }

    // Drains the values stored in self.next, and maps them to futures where they can be collected.
    fn transition_next_to_futures(&mut self) -> Vec<impl Future> {
        println!("transitioning futures: next: {:?}", self.next);

        // TODO: using std::mem::replace here results in weird errors. Are we not supposed to move
        // the data behind an Arc???

        // let next = std::mem::replace(&mut self.next, Arc::new(Mutex::new(HashSet::new())));

        let mut next = self.next.lock().unwrap();
        next.drain()
            .map( |node| {
                self.process_node(node).boxed()
            })
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
    let mut tracker = NodeTracker::default();
    tracker.add_node('a');

    let mut pending_futures = Vec::new();

    loop {
        let mut next_futures = tracker.transition_next_to_futures();
        pending_futures.append(&mut next_futures);
        println!("pending_futures.len(): {}\n", pending_futures.len());
        if pending_futures.is_empty() {
            break;
        } else {
            // Block until one of the futures is ready:
            let (_item_resolved, _index, pending) = select_all(pending_futures).await;

            println!("\npending.len(), after race: {}\n", pending.len());
            pending_futures = pending;
        }
    }

    let total = tracker.rewards.value();
    println!("total: {}", total);
    assert_eq!(total, 4250);
    Ok(())
}

async fn fetch_node(
    node: Node,
) -> Result<Response> {
    let resp = reqwest::get(&format!("{}/node/{}", URL, node))
        .await?
        .json::<Response>()
        .await?;
    Ok(resp)
}
