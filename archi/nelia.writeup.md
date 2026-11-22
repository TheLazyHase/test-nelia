Sampling transmission chain
===========================

I - Goal
--------

The goal is to

II - Hypothesis
---------------

* In the first phase we don't distinguish between samples that have been accepted at the lab for analysis and the one whose analysis is finished

* sample collections points can print labels to quickly put them on samples or boxes

* samples may have problems that the coursiers cannot see but that the lab staff can

* We may have to expand to additional laboratories in the future. No need to have specific features for that yet, just be aware that the laboratory may not be unique.

* We accept that the service is entirely down if there is no connections possible to central server. This may be remediated later

III - Workflow
--------------

The actors are :
* the patient. Note that we don't distinguish the patient themselves from people acting on his behalf (the doctor who prescribed the exam, the lab worker that book the appointment)
* the technician, which is a person at the sample collection point that take the sample, then put it in a transport box
* the coursier, who take the transport box and ship them to destination
* the lab staff, who receive the transport boxes.

1. Patient book a sample taking. Our system create a recording for that sample. The technician get a printed label with a sample identifier
2. Technician take the sample from the patient, and put the label printed previously on it.
4. Technician put it into the transport box (may be merged with 3 depending on physical factors, but currently I take the hypothesis it's an entire step), and check in our system that it's in the boxe
5. coursier check all the samples in the box in our system, and transport it to the lab
6. coursier reach the lab, lab staff open the box, check every individual sample for problems, and start the analysis.

IV - Irregularities and incidents
--------------------------------

Regardless of how well trained everyone is and how bugless the code is, strange, unexpected events can happen, and it's important to act on them, but also to avoid paralyzing the system by being overzealous. I propose a two-tier level

* Irregularities are when something unusual but still within normal operating parameters happen, like a sample starting its travel with a coursier and finishing with another. It's important to record them to detect if something unusual is happening or simply to be able to root out bad habits, and if a lot of them happen to a single sample or a lot more of them happen in a day, it's a sign we need to investigate

* Anomalies are  unexpected events that should leave alarms and immediatly call a human to investigate. For example, the same sample being preleved twice, additional samples being found, etc.

For both kind, discussion with the product and the end users are important during implementations. Treat any mention of something being an irregularity or an anomalie as being my best guess and open to discussions, not a directive.

Here is the list of irregularities :
* sample is booked but never actually taken (every day some peoples will have problems preventing the actual sample taking, we just want to be aware of how many of these happen)
* same box is checked by coursier several times (the hypothesis is that they had to reshuffle their loads, or they had a vehicle failure and gave the box to another coursier)
* same sample is checked in at the lab twice (if a lab staff inadvertantly check the same item twice, we want to log in but not paralyze the whole system to investigate)
* coursier that check the box out is not the last one to have checked in (meaning they didn't correctly check in their samples to the new coursier)

Here is the list of anomalies :
* technician declare putting a sample in a box but that sample have already been checked in by a coursier
* technician take the same sample two time. (if the sample got destroyed by carelessness, correct workflow is to create a new sample entry in our system entirely)
* missing or duplicate samples

On top of all these automatic irregularities and anomalies, every time a sample is checked the system allow to fill a field "notes" with a choice of it being an annotation, an irregularity, or an anomaly.

V - Security
-------------

There's two relatively different attacker types we need to be aware of.

The first are data thieves, whose goal will be to extract as much information from the system as possible. Since we deal with medical informations, preventing data theft and mitigating it when it happen is important. This lead to a number of measure :
* strong decoupling of the APIs used by laboratories, coursiers, and sample collections points, so if someone break one of these authentication he only access to a limited range of data
* making sure any API only return relevant informations, and not anything supplementary. For example, coursiers don't need to know the names of the owner of the samples they transport.
* the needed back-office API that have acces to everything need to have every access logged in an immutable way and have a secure login. Secondary measures like geofencing could be added later.

The second ones are tamperers, people who for various reason would want to tamper with analysis results, by switching two samples typically. Most of the work here is down to training and physical security, but there's several measures to help with this :
* have an immutable event journal for samples, so even if a tamperer would modify the main database we can rebuild it from the journal.
* use non-sequential UUID that is not purely time-based so an attacker cannot guess the ID of a sample
* appropriately graph samples states to more easily discern unexpected patterns
* use idempotency key on every status change, to trim down anomalous events.

VI - Tables
------------

We will leave apart the identity provider for labs, collection points, and their staff (who are explicitely external SSO) and the one for coursiers and patients (unaddressed in the briefing, may very well be an external SSO too, and a standard thing)

1. samples

This is the main table used by the system, each describing a single sample and where it's supposed to be.
Columns :
- ID, an unique non sequential identifier as primary key.
- states :
 * booked
 * taken
 * in box
 * in transit
 * accepted
 * archived
 * incident
 * investigated
 * cancelled
 the first 6 are normal state, with archival rules needing to be defined later. The last 3 is related to anomalies ; if a sample have an anomaly and cannot be treated, it's put in incident, and admins can put it in cancelled to effectively archive them as failed, or investigated, to signal someone is investigating on it. We may also add rules to push from incident to cancelled automatically in some cases
- notes, who is a json where each successive note is appended together with whether the reporter consider this normal, an irregularity, or an anomaly
- irregularity_amount, increased each time a note is an irregularity or the system detect one
- anomaly_amount, increased each time a note is an anomaly or the system detect one
- patient_id
- technician_id, can be null in booked state
- collection_point_id, can be null in booked state
- coursier_id, can be null until state is "in transit"
- laboratory_id, can be null until state is "accepted"
- standard columns like created_at, updated_at
This table will have a lot of keys, entire lists to be determined but it will need to be searcheable by ID, state, coursier_id, among other

2. sample_events

This table is for better investigations when problems happen. Neither the API nor the backoffices are meant to ever delete or modify lines in this table.

Every time an operation on a sample is done, after updating the main table, we insert a line in this table with the same fields as samples, but also :
* previous_state and new_state instead of a single state. Note that they may be identical (the aforementioned "a box is checked in by a new courier" for example)
* reporter_id, which is the ID of the reporter
* note only contain the current note (to save space and be more readable)

This table is not destined for heavy use and will not have a lot of indexes.

VII - Software architecture
---------------------------

Some discretion is allowed to the implementer. All routes need to check if the user who is using the route is allowed to.

There will be 3 APIs :
* collection point API, that require being identified to the collection point SSO. Routes :
 - GET /samples/me who report all samples made in that laboratory who are not archived. Only fields relevants are returned.
 - POST /samples/book who create a new sample. As above, if the note is an anomaly it put the state to incident
 - POST /samples/event who create a new note for the sample without changing state.
 - POST /sample/take who push the sample from booked to taken state and assign the logged in user as technician_id. User can add a note
 - POST /sample/boxe who push the sample from taken to in box state. User can add a note.
*(at the discretion of implementer, the POST routes can be merged and replaced as an action on the route)*

* coursier API, that require being identified to the coursier identity provider. Routes :
 - POST /samples who check in the given sample to that coursier. Coursier ID is assigned to the sample. User can add a note. It can be used by another coursiers, in case they need to pass the sample to someone else (see irregularities and anomalies)
 - GET /samples/me who report all samples that that coursier have checked in, with relevant fields. This is helpful if a coursier lose a sample and need to find which one to declare lost.

* lab API, that require being identified to the lab SSO. Routes :
 - GET /samples/me who report all samples that that the lab expect or hold with relevant fields. "expect" mean "have been checked in by a coursier in destination to the lab" (state "in transit")
 - POST /samples/accept who accept the given sample to that coursier. User can add a note.
 - POST /samples/event who allow to add a note to the sample. (possibly an irregularity or anomaly)

* back office, used by admins to investigate problems.
 - GET /samples, with full filtering and all needed field. Particularly they need
 - GET /samples/history who return all the events of the table sample_events related to that sample.
 - POST /samples to modify a sample any way needed. Note that this will be logged in sample_events.
This should allow for any needed workflow for admins who want to investigate an event.

To reiterate, each POST action should log into sample_events on top of altering the samples table.

I advice to have for each API different handlers and services but same repository ; different handlers and service mean less risk of failing to correctly clamp results to what the user is allowed to see.

VIII - Load
-----------

The system will have low load at this stage (~10 000 rows per days, each who normally need ~10 updates) but reliability will be important. Standard databases should do the job as long as indexes are properly set.

To further avoid problems, archived lines should be sent to cold storage regularly, then ultimately deleted when old enough.

There may be a need for iocaine or similar anti-bot solution.

IX - Expansions
----------------

* send archived samples to cold storage regularly
* automatic rules to create irregularities or incident, or to put samples in archived or cancelled without human intervention
* use of a queue system like kafka instead of calling the API directly. Not currently needed given the expected volumes and centralized system (only one laboratory). This also may allow a better offline degradation, albeit legal requirements may make that undesirable anyhow.
* expand to several central laboratories, which will require researchs about lab assignments
