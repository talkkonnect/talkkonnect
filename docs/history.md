# History and Background of talKKonnect

## Where did talKKonnect come from?

This project is a fork of [talkiepi](http://projectable.me/) by Daniel Chote which was in turn a fork of [barnard](https://github.com/layeh/barnard) a text basedmumble client.
talKKonnect was developed using [golang](https://golang.org/) and based on [gumble](https://github.com/layeh/gumble) library by Tim Cooper.

Most Libraries are however heavily vendored (modified from original). You will need to get the vendored libraries from this repo. Talkkonnect has implemented using the later specs the mumble protocol, so please use the talkkonnect vendored libraries (gumble) for building talkkonnect. Using original gumble library has does not have channel listening features and the build will fail because of missing functions mapped to the vendored version of the gumble library.

[talKKonnect](http://www.talkkonnect.com) was developed initially to run on Linux SBCs. The latest version can be scaled to run all the way from ARM SBCs to full fledged X86 servers.

To compile on X86 archectures you would need to revert back to Tim Cooper's version of GOOPUS (Opus) since the older build supports x86 processors.

Raspberry Pi 2B,3B,3A+,3B+,4B,400,Zero 2W, Orange PI Zero H2 Chip targets have all been tested and work as expected.

For the The enthusiast or those who want to test the features of talkkonnect the newly released Raspberry Pi Zero Version 2W used with a respeaker hat and external speaker is the perfect candidate for test driving talkkonnect. With this hardware no soldering required and ready made images are available for burning and testing.

### Why Was talKKonnect created?

I, [Suvir Kumar](https://www.linkedin.com/in/suvir-kumar-51a1333b), created talKKonnect for learning and fun. I missed the younger days making homebrew CB, HAM radios and talking to all
those amazing people who taught me so much. My HAM Radio Call signs are HS1FOS/E25OSW for Thailand and KK7HMK (Extra Class) for USA.

Living in an apartment in the age of the internet with the itch to innovate drove me to create talKKonnect. I did it as a hobby to learn, so in no way am I a professional programmer, however talkkonnect is very stable and running in production for mission critical systems all over the world and is production ready. That being said brace yourself for some code from a self taught amateur programmer some parts of talkkonnect are not ideal.

I have tried to make the talKKonnect source code readable and stable to the best of my ability. Time permitting I will continue to work and learn from all those people who give feedback and show interest in using talkkonnect.

[talKKonnect](http://www.talkkonnect.com) was originally created to have the form factor and functionality of a desktop transceiver. With community feedback we started to push the envelope to make it more versatile and scalable as you can see from the rich feature list. We also later added announcement and PA abilities to make talkkonnect support IP-Speaker functionality.

### A Video Introduction to talKKonnect
<iframe width="560" height="315" src="https://www.youtube.com/embed/nLmHM48SqFs?si=ApowQ_h-F449Zq1Y" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>