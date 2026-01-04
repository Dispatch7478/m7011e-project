## Ethical analysis:

### Sensitive data handling:

Keycloak is excellent for user authentication as it provides great security. Allowing keycloak to store passwords etc gives greater protection than if we were to do so. It also allows us to have features such as MFA and password requirements to ensure that the user is as safe as possible. We can also use RBAC with keycloak to ensure that users can only access data that they are authorized to. 

Something outside of keycloak which can be seen as positive is that each service has its own database. In the case that one database would be compromised in an attack the others would still be safe.

As for what is missing we are slightly unsure, we believe that most of the sensitive data is handled in a way that protects the users well. One thing is that there still could be some potential exploits with headers in the api gateway as the email possibly could be intercepted.
 
### Privacy implications:

In our current system there are quite a few faults when it comes to privacy implications. 
What we do ensure is that no excessive data is being taken and only store that which is necessary for the system to work and that none of the data is shared or distributed in any way outside of the web app. 

We do not have ways for users to be more private and being able to hide their activity to an extent could be more appropriate. 

### Societal impact:

The platform helps people get together in one place where they can enjoy their competitive hobbies. A great way to build communities and social engagement which can otherwise be difficult yet important.
