// Package mailbox provides a mailbox communication pattern where different actors
// can send messages to each other, even if they do not have a direct line to
// each other
//
// # Introduction
//
// There is no implict order of delivery of messages, because an actor might be
// connected to one or more rack instances at the same time and there is no requirement
// that each rack instance talks to each other.
//
// There is a possibility that the same message might be delivered more than once
// to the same actor (replication or durable storage), it is up to the actor
// to de-duplicate such messages.
//
// To help with that, each message contains a field to indicate when it first arrived
// at a specific rack, and senders MUST also provide a time stamp when they send messages.
// During subscription, actors can specify filters to reduce how much traffic flows from
// the rack to the actor.
//
// Messages might be of two types:
//   - Request/Replay pair, the rack will not control if a given pair is valid, therefore
//     an actor might receive a Replay for a message that was never sent. Actors should maintain
//     internal state to handle such scenarios. As well as timeout replies that were never received.
//   - Notification where no response is expected on the other hand
//
// The mailbox package does not implement any special process for those types of messages,
// although it DOES specify some of those fields. A message for a unknown address might be
// silently dropped if there are no registered consumers for it.
package mailbox
