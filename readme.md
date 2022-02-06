# Blockchain aircraft management

The following is copied from my devpost entry: 

## Inspiration
Flying is one of the safest methods of transportation available to people today, however errors and structural failures still do happen.  When a catastrophic tragedy like this occurs it is important for people to learn from their mistakes and analyze the cause of a crash.  Recent advancements in technology have allowed people for people to create in flight recorders that are capable of logging hundreds of different messages, equipment details and flight metrics to make determining the cause of a plane crash easier than it would normally be.  Despite the leaps in modern technology that allows the in flight recorder work, there is one fatal flaw in it's design, there is only one logger, and thus only one copy of the data it holds.  This single copy of the data can easily be corrupted or destroyed during the high velocity impact of a crash.  Despite the best efforts of engineers, of the 684 fatal commercial aircraft crashes since 1959 (Huffpost) there were 59 in flight recorders lost during the aircraft's impact with the ground (Wikipedia).   My project seeks to solve this problem...

## What it does
My project seeks to solve this problem by forming a decentralized network of blockchain nodes throughout every piece of an aircraft that allows data and communications about in flight performance to be redundantly stored and integrity checked in every computerized system of the aircraft.  This means that if any one part of the aircraft is destroyed or damaged beyond repair, another part will hold a redundant copy of all of the information.  With a system like this, in the event of a crash even if nothing on an aircraft other than one of it's hydraulic pumps stays together in a crash, all flight records will remain available for investigators.

## How I built it
The back-end of my project is a Go program that manages a blockchain designed to ensure data integrity and preservation.  The entire program has no external dependencies outside of the Go standard library.  The Go program front-ends to a HTML and CSS web interface that I developed in order to make use and demonstration of the process much easier.

## Challenges I ran into
The main challenge that I ran up against during this hack-a-thon was my lack of knowledge on blockchains.  Prior to 1 day before the competition I had only a cursory knowledge of how blockchains worked.  In order to prepare I spent the entire day before reading about the technical inner workings of different blockchain networks like Bitcoin and Ethereum in order to gain a better understanding of the fundamentals of blockchains.  Despite the research I did, the lack of study time I had meant that there were a few holes in my logic regarding blockchains, this caused me to spend extra time researching when I could have been coding.

(To be clear the only thing I did before the start was research, no code was produced!)

## Accomplishments that I am proud of
I am proud that I was able to learn enough about blockchain technologies during my small span of study time to create one of my own, especially one that is different from almost all other ones in existence.

## What I learned
The planning and execution of this project has taught me an immense amount about blockchain and their technologies.  Before I had the idea for my project, I knew very little about cryptocurrency and other types of blockchain technologies.  In the past few days I learned more than I could have ever thought possible.  I now understand many of the concepts that allow cryptocurrencies to function as well as the mechanisms and algorithms that allow them to function.  I can now confidently say that I know how blockchains work and advise others in matters involving them.

## What's next for blockchain aircraft management.
Because of the large barrier to entry present into the aerospace field as well as how much lifespan air travel companies tend to squeeze out of their previous investments, it is unlikely that we will see blockchain technology reach large scale implementation into aircraft for many years.  However, his doesn't mean that the tech has no value.  The exact same software that I have demonstrated can be used in a plethora of different vehicles and complex system in order to reap the advantages that I have outlined.  In the future I think that this tech will first reach widespread adoption in other fields like space exploration, automotive and even bio-medical science.  The opportunities and applications are truly endless, aircraft were only a means of demonstration.
t

All code (c) sckzor 2022
