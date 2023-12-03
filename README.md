## Disclaimer
This is unlikely to be interesting to anyone but me.

I wrote it to answer a concrete question I had (is the desired
schedule feasible?). It's not meant to be general-purpose or even
particularly educational.

## Motivation
Create a schedule for a league with:
- 6 teams
- 6 days of play
- 3 time slots during each day of play, in which two teams can play a match.

## Requirements
- Every team plays every other team at least once, and at most twice.
- No team plays twice in a day, and no team has a bye day.
- Every team plays twice in each of the 3 time slots during the schedule.

## Implementations
- `main.go` solves the problem using `LPSolve` (via `golp`) by translating the requirements into an integer linear program. Example output is:
    ```
    BC AB AD CD EF EF 
    AE CF CE AF BD BD 
    DF DE BF BE AC AC 
    ```

- `main2.go` solves the problem using brute-force recursive search. It hard-codes the first day's matches to make it easier to eyeball the results. Example output is:
    ```
    AB AC BF CF DE DE
    CD BE CE BD AF AF
    EF DF AD AE BC BC
    ```
