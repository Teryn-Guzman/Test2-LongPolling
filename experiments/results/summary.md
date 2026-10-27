# Assessment 2 measured comparison

Every row below is generated from the supplied client/server logs. Blank detection values mean the terminal commit could not be joined. Notes show only exceptions and skipped-version counts; detailed version records remain in the raw logs.

## Input check

| Log | File | JSON records read | Non-JSON lines skipped |
|---|---|---:|---:|
| client | experiments/raw/client.jsonl | 1788 | 0 |
| server | experiments/raw/server.jsonl | 3505 | 0 |


Skipped lines are excluded from the analysis only; the raw files are not modified.

Runs excluded at the operator's request (raw log unchanged):

| Job | Condition | Mode | Rep | Client events |
|---|---|---|---:|---:|
| e9c04638-e6f5-4b3a-934f-dee8b1eee02e | A | short | 1 | 38 |
| 28d43985-136f-4b46-9db7-46455e1ce3c8 | A | long | 1 | 16 |

## Workload check

All 30 run(s) match the server flags and latency note expected for their condition label.

Artificial delay is the configured quiet and image delay plus the measured simulated phase. Real processing is processing_ms minus that; it still includes claim-to-finish overhead.

## Runs

| Run | Condition | Mode | Rep | Job | Requests | Detection (ms) | Processing (ms) | Artificial (ms) | Real (ms) | Versions | Notes |
|---|---|---|---:|---|---:|---:|---:|---:|---:|---:|---|
| 01629ec4-3635-4b08-8ef0-fc8af68b588e | C3 | long | 3 | ec1fff49-0779-4a8b-b9e9-68c0a6f9327e | 48 | 267 | 10222 | 10000 | 222 | 47 | 5 intermediate version(s) skipped |
| 06cb07de-78a2-4722-8bdc-e48eb70d600a | C3 | long | 2 | a8f675ff-d56e-40c9-bf39-254c04c94dd1 | 49 | 359 | 10219 | 10000 | 219 | 48 | 4 intermediate version(s) skipped |
| 0ab9a916-4f2d-45be-aefb-f05b5a225698 | C2 | long | 1 | 318fdd55-98d2-4050-90ba-3e91b1339fc2 | 54 | 11 | 10248 | 10001 | 247 | 53 |  |
| 168fe210-5471-49a3-8e9f-4019feca357e | A | long | 3 | 0287b0bf-866d-4b29-b84a-e9d5bd230c96 | 6 | 12 | 30291 | 30000 | 291 | 4 |  |
| 1958a209-40dd-40b9-8585-b565777708ab | A | short | 3 | 826640ae-e335-431b-8e2a-dae7ee3c2208 | 31 | 199 | 30206 | 30000 | 206 | 2 | 2 intermediate version(s) skipped |
| 200326d7-16da-4650-b043-c7c84a8c8be6 | C2 | short | 1 | 5b64c265-d65f-4111-8cc3-459b975cb041 | 12 | 789 | 10214 | 10001 | 213 | 11 | 42 intermediate version(s) skipped |
| 20061711-ac99-4283-9ef2-f17ebbf2bcca | A | long | 1 | 2c52f64c-b821-4e4f-a680-7b6371ddcefe | 6 | 3 | 30288 | 30000 | 288 | 4 |  |
| 3ab87b43-b82c-434a-8ed8-e649219a2c7d | B | short | 3 | 8292f8ed-d3e1-4613-b905-650e68b60cb3 | 2 | 792 | 220 | 0 | 220 | 1 | 3 intermediate version(s) skipped |
| 3bd7ae63-4544-476d-adfc-4012c68ae40b | C2 | long | 3 | 24f05f52-7234-4258-9767-2b0cddf117a6 | 54 | 2 | 10226 | 10000 | 226 | 53 |  |
| 3c8f1fe9-8ae0-48e2-9ad4-a7a2d6db399c | C1 | long | 2 | e6fd278a-4c74-4635-a03f-41630e96faa7 | 9 | 10 | 10261 | 10000 | 261 | 8 |  |
| 476b2785-0497-4263-b426-01ec85a58cb9 | C1 | short | 2 | eeb9b09d-67fd-4994-8971-aea866b34586 | 12 | 735 | 10228 | 10001 | 227 | 6 | 2 intermediate version(s) skipped |
| 527669d3-e342-481f-9e86-0c8c57225554 | C3 | short | 3 | a5b7a0d3-0d8c-47f0-a5e4-0923f8fe1dbf | 10 | 1036 | 10227 | 10001 | 226 | 9 | 43 intermediate version(s) skipped |
| 52f545cc-1ff1-4bdb-80ef-71bd56ee9c8f | C1 | long | 1 | 5ffca700-0636-4acc-94eb-7c9c9524445a | 9 | 4 | 10247 | 10000 | 247 | 8 |  |
| 5ccad4f0-9333-432b-8bb4-b2493191f361 | C1 | short | 3 | dd801e32-4268-49a5-a5c5-54bd310d4bcc | 12 | 802 | 10269 | 10000 | 269 | 7 | 1 intermediate version(s) skipped |
| 6451bead-5518-47a6-8cc1-e1d1d62efe0f | C3 | long | 1 | 259e7926-35ad-4af2-9f01-a5e3edda738d | 49 | 309 | 10240 | 10000 | 240 | 48 | 4 intermediate version(s) skipped |
| 6911c840-cc92-4d08-b8f4-d30337b068ee | A | short | 1 | a5b46756-48b1-4d79-889a-fd3b7aa1ebdb | 31 | 123 | 30235 | 30000 | 235 | 2 | 2 intermediate version(s) skipped |
| 768a5652-29e6-401f-80a2-5d5982d1da9e | A | short | 2 | 894da8df-149c-468f-94de-18f477d1a597 | 31 | 99 | 30221 | 30000 | 221 | 2 | 2 intermediate version(s) skipped |
| 89fc4c34-25bc-457a-a7a4-243922bd0600 | C2 | long | 2 | 9d61c548-785d-4c4c-9411-979f8658c6e3 | 54 | 3 | 10222 | 10000 | 222 | 53 |  |
| 91013a06-1f57-46e0-9ca3-75e43db8d206 | A | long | 2 | 090fbe3f-8544-4590-aa67-f2772950ba03 | 6 | 5 | 30278 | 30000 | 278 | 4 |  |
| 91ea1cf3-4266-4f35-b8d9-04fce072138e | C1 | short | 1 | 5a3b5cc9-837d-40ea-a4b0-fb563f19706d | 12 | 976 | 10230 | 10000 | 230 | 7 | 1 intermediate version(s) skipped |
| 94015154-b073-46c8-b0eb-56dcc8d487dd | C2 | short | 2 | 0fcaa89e-f30e-4b4a-b249-efdba213bf3e | 12 | 744 | 10218 | 10000 | 218 | 11 | 42 intermediate version(s) skipped |
| a6d16ffd-8ce6-4c84-8dea-5ff822fbdf49 | C2 | short | 3 | ccc4b386-d9fa-4055-bb28-25848e006da3 | 12 | 982 | 10225 | 10000 | 225 | 11 | 41 intermediate version(s) skipped |
| c974dbf5-0d8b-4dd1-99be-830525946f4c | C3 | short | 1 | c72133b4-cbcf-408d-8cd7-1a939db9d37f | 10 | 1060 | 10216 | 10000 | 216 | 9 | 43 intermediate version(s) skipped |
| cb8cef20-2a50-41e6-ba55-e08371da2065 | C1 | long | 3 | 76aa1a81-aab6-46ea-9162-e3d511091cf9 | 9 | 5 | 10301 | 10001 | 300 | 8 |  |
| d2612352-ae5d-4ecf-9430-e45341a5da37 | C3 | short | 2 | cbb77445-a9f8-4797-98d2-f80bf75918d1 | 10 | 1017 | 10220 | 10000 | 220 | 9 | 43 intermediate version(s) skipped |
| e31ee812-fe4b-4b42-94d7-cf2e957fc7f9 | B | long | 1 | 821ee60f-c127-40da-bb72-991abbd1b6a2 | 5 | 2 | 211 | 0 | 211 | 4 |  |
| ee77cdc3-b5c5-4bbf-884a-6db036fa4df1 | B | long | 2 | 192c2d46-512e-487c-9d51-7fe726208a13 | 5 | 2 | 222 | 0 | 222 | 4 |  |
| eecc11a2-e495-4671-a996-8a61409c6ecd | B | long | 3 | 68cc30a5-9c50-4df9-b0d5-af16e694662f | 5 | 4 | 217 | 0 | 217 | 4 |  |
| f3d75163-b9b6-412d-90d9-15ca67391b5e | B | short | 1 | 93ebde86-8ca8-47a6-8b20-d0122c47be66 | 2 | 632 | 216 | 0 | 216 | 1 | 3 intermediate version(s) skipped |
| f8f2a804-02b8-4def-b461-987b29a94fdb | B | short | 2 | fd19f378-75f6-4c5c-8d81-ee2d30670764 | 2 | 629 | 204 | 0 | 204 | 1 | 3 intermediate version(s) skipped |

## Median (minimum–maximum) by condition and mode

| Condition | Mode | Requests | Detection delay (ms) |
|---|---|---:|---:|
| A | short | 31 (31–31) | 123 (99–199) |
| A | long | 6 (6–6) | 5 (3–12) |
| B | short | 2 (2–2) | 632 (629–792) |
| B | long | 5 (5–5) | 2 (2–4) |
| C1 | short | 12 (12–12) | 802 (735–976) |
| C1 | long | 9 (9–9) | 5 (4–10) |
| C2 | short | 12 (12–12) | 789 (744–982) |
| C2 | long | 54 (54–54) | 3 (2–11) |
| C3 | short | 10 (10–10) | 1036 (1017–1060) |
| C3 | long | 49 (48–49) | 309 (267–359) |

## Run-count check

Observed 30 runs; required minimum is 30 (3 repetitions × 5 conditions × 2 modes).
