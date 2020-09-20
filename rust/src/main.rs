#![warn(clippy::all, missing_debug_implementations, rust_2018_idioms)]

use futures::future::{select_all, BoxFuture};
use futures::FutureExt;
use serde::Deserialize;
use std::collections::HashSet;
use std::sync::mpsc::{sync_channel, Receiver, Sender};
use std::sync::{mpsc::channel, Arc, Mutex};
use std::thread;
use std::thread::JoinHandle;
use tokio::macros::support::Future;

type Error = Box<dyn std::error::Error>;
type Result<R, E = Error> = std::result::Result<R, E>;

const URL: &str = "https://graph.modulitos.com";

type Reward = u32;

enum RewardMessage {
    Increment(Reward),
    Terminate,
}

struct RewardCounter {
    thread: JoinHandle<Reward>,
}

impl RewardCounter {
    fn new(receiver: Receiver<RewardMessage>) -> Self {
        let thread = thread::spawn(move || {
            let mut total = 0;
            loop {
                use RewardMessage::*;
                match receiver.recv().unwrap() {
                    Increment(reward) => {
                        total += reward;
                        println!("total: {}", total);
                    }
                    Terminate => {
                        break;
                    }
                }
            }
            total
        });
        Self { thread }
    }

    fn join(self) -> Reward {
        self.thread.join().unwrap()
    }
}

type Node = char;

/// Handles the nodes
/// TODO: consider leveraging a mpsc channel here instead of Arc/Mutexes?
struct NodeTracker<'a> {
    // Nodes that have already been fetched and processed.
    visited: Arc<Mutex<HashSet<Node>>>,
    // The next nodes that we want to fetch:
    next: Arc<Mutex<HashSet<Node>>>,
    // The nodes currently being fetched, represented as Futures
    pending_futures: Vec<BoxFuture<'a, ()>>,

    // Used for communicating the rewards for further processing:
    reward_sender: Sender<RewardMessage>,
}

impl<'a> NodeTracker<'a> {
    fn new(sender: Sender<RewardMessage>) -> Self {
        Self {
            visited: Arc::new(Mutex::new(HashSet::new())),
            next: Arc::new(Mutex::new(HashSet::new())),
            pending_futures: vec![],
            reward_sender: sender,
        }
    }

    fn add_node(&mut self, node: Node) {
        let visited = self.visited.lock().unwrap();
        let mut next = self.next.lock().unwrap();
        if !visited.contains(&node) && !next.contains(&node) {
            next.insert(node);
        }
    }

    fn process_node(&self, node: Node) -> impl Future<Output = ()> + 'a {
        let visited = self.visited.clone();
        let next = self.next.clone();
        let reward_sender = self.reward_sender.clone();
        async move {
            let resp = fetch_node(node)
                .await
                .expect("HTTP request to fetch and deserialize the node failed!");

            println!("resp from '{}': {:?}", node, resp);

            reward_sender.send(RewardMessage::Increment(resp.reward));

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
    // returns whether the pending futures are empty.
    fn transition_next_nodes_to_futures(&mut self) {
        println!("transitioning futures: next: {:?}", self.next);

        // TODO: using std::mem::replace here results in weird errors. Are we not supposed to move
        // the data behind an Arc???

        // let next = std::mem::replace(&mut self.next, Arc::new(Mutex::new(HashSet::new())));

        let mut next = self.next.lock().unwrap();
        let mut pending_futures = next
            .drain()
            .map(|node| self.process_node(node).boxed())
            .collect::<Vec<BoxFuture<'a, ()>>>();

        self.pending_futures.append(&mut pending_futures);

        println!("pending_futures.len(): {}\n", self.pending_futures.len());
    }

    async fn wait_for_next_node(&mut self) {
        let pending_futures = std::mem::replace(&mut self.pending_futures, vec![]);
        let (_item_resolved, _index, mut pending) = select_all(pending_futures).await;

        println!("\npending.len(), after race: {}\n", pending.len());
        // reset our pending futures to be the ones that are remaining:
        self.pending_futures.append(&mut pending);
    }

    async fn run(&mut self) {
        loop {
            self.transition_next_nodes_to_futures();
            if self.pending_futures.is_empty() {
                self.reward_sender.send(RewardMessage::Terminate).unwrap();
                break;
            } else {
                // Block until one of the futures is ready:
                self.wait_for_next_node().await;
            }
        }
    }
}

#[derive(Deserialize, Debug)]
struct Response {
    children: Vec<Node>,
    reward: Reward,
}

#[tokio::main]
async fn main() -> Result<()> {
    let (reward_sender, receiver) = channel();
    let mut tracker = NodeTracker::new(reward_sender);
    tracker.add_node('a');

    let worker = RewardCounter::new(receiver);
    tracker.run().await;

    let total = worker.join();

    println!("final total: {}", total);
    assert_eq!(total, 4250);

    Ok(())
}

async fn fetch_node(node: Node) -> Result<Response> {
    let resp = reqwest::get(&format!("{}/node/{}", URL, node))
        .await?
        .json::<Response>()
        .await?;
    Ok(resp)
}
