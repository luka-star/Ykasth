# Take-Home Technical Assignment

**Expected time:** 2–5 hours  

This task is meant to reflect the kind of work you’d do with us:
integrating external data, structuring code clearly, and reasoning about system design.

We are not testing memorization, syntax trivia, or production-grade polish.

---

## Context

You are building a small service that:

1. Fetches data from an external API
2. Processes or transforms the data
3. Publishes it to a message queue (simulated)
4. Stores the result in a simple in-memory structure

This mirrors our real system, just at a much smaller scale.

---

## Task

Build a small command-line application that does the following:

### 1. Fetch external data periodically (polling is fine)

Use any public API you like (e.g. weather, crypto prices, placeholder APIs), but respect rate limiting.

If you want something closer to our domain, you can use:
- https://www.energidataservice.dk/guides/api-guides  
  (ImbalancePrice dataset)

This is optional — choose what you’re most comfortable with.

---

### 2. Abstract the data source

Structure your code so that:
- The data source is hidden behind a clear abstraction (interface / base class / equivalent)
- The rest of the system does not depend on the concrete API implementation
- The implementation do not need to be over-engineerd

The goal is to show how you think about separation of concerns, not to use a specific pattern or language feature.

---

### 3. Process the data

Transform the data in a meaningful way, for example:
- Filtering fields
- Converting units
- Aggregating values
- Producing a small summary

Keep it simple and explain your choice.

---

### 4. Simulate a message queue and publish data
For example: an in-memory channel/queue that other parts of your code could consume from

It’s enough to simulate by:
- writing to an in-memory queue/channel
- Consume from the queue and printing structured messages to the console

This is about showing an understanding of *why* systems use message passing.

---

### 5. Store results

Store the processed data in memory (array, map, struct, etc.).

No database setup required.

---

### 6. Optional extras (only if you have time)

You may implement one or more of the following:
- Detect when data is new and only publish when it changes
- Add a small unit test or two
- Make the application configurable (config file or environment variables)
- Add basic logging with different log levels
- Add retry / backoff
- Make the fetch interval configurable.

---

## Requirements

- Use any language you are comfortable with (If you’d like to align with our stack, we mainly use Go and TypeScript, but this is optional.)
- Keep responsibilities clearly separated
- Handle basic error cases (e.g. failed API call)
- Make reasonable assumptions — document them
- The application should keep running and publish to message queue until stopped.


---

## What to Include

Please include a short README explaining:
- How to run the project
- How you structured the code
- Any assumptions or trade-offs you made

Add comments where it helps explain *why* something exists.

---

## Review

After submission, we’ll go through your solution together and discuss:
- Your design decisions (when you used abstraction and when you did not)
- Possible extensions or changes
- Failure cases and improvements
- ...

---

If anything is unclear, feel free to ask questions.
