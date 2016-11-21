# Cryptonite
<p align="center">
  <img src="https://wiki.staypirate.org/images/Cryptonite.jpg"  height="150" width="150" alt="Cryptonite logo"/>
</p>
## Abstract
This malware is intended to be a backdoor for the commonest OSes. It's been written in GO to easy cross-compiling static binary for different architectures and OSes. That also simplify future deployments.

It's composed of two parts: *stager* and *backdoor*.
 * The first one is the one that will be injected and executed into the victim machine exploiting a vector. The main task of this one is to understand the environment, what kind of OSes and architecture it's running on, then download and deploy the right architecture backdoor from a specific server.
 * The second one is the real malware which implements the functions listed below.

## Functions
 * [Certificate pinning with double couple keys] (#certificate-pinning-with-double-couple-keys-notes)
 * Proxy (ssl tunneling?)
 * POST and GET for files and directories (sshfs)
 * Keylogger
 * Grab saved password in the browser
 * Grab LUKS master key (with [t4ub](https://github.com/StayPirate/t4ub))
 * Reverse shell
 * Take screenshot
 * Connection to C2s (Command and Control Server)
  > C2s as hidden service (connect zombies through [TOR](https://gitweb.torproject.org/tor.git/))

 * Disinstallation

Each functions should works on each target OSes.

## Target OSs
 1. Windows 7/8/8.1/10
 2. GNU/Linux
 3. Mac OSX

## Double certificate pinning
The bootmaster have to create his own CA, that can be done by [gen-keys.go](gen-keys.go), at first setup.

The certificate of this CA will be encoded into clients. That ensure zombies will connect only with certified C2S.

### Private C2S's key
To deploy a Command and Control Server, we have first generate it's own certificate then sing it by CA

Of course, when we are going to deploy a C2S we have to deploy too the private key and relative certificate with a valid signature.

Could happen sometimes we want deploy the C2S into a *not our* VPS (maybe owned by someone who doesn't know that we are using his server)... then the problem is:
> What if the real owner find our C2S and keeps a backup of keys?

He could use/abuse/steal our botnet. That's why we have a *couple of certificates* encoded in each clients.

### Bootmaster (BM) keys
At first setup, using [gen-keys.go](gen-keys.go), there will be generated BM's pub and pvt keys.
> *Note*: The pubkey is not wrapped into a x509 cert, but just a pubkey. Like ssh-keys.

The pubkey will be encoded into clients as happens for the CA's certificate. Meanwhile the pvtkey, should remain in the BM's computer.

The pvtkey will be used to sign every command sendeds from BM to zombies. Those will check signature before execute each command they receive.

A copy of the BM's pubkey will be encoded into every C2S too, hence it can decides if forward the incoming command to the zombies by checking signature first.

But there is a problem here:
> What if the server real owner keeps a copy of one valid command that we sent to the zombies, and decides to replay it infinite times?

He could DOS our botnet or iterate attacks (like DDOS) that we had already stopped.

### One-Time-Command

Each command should have different signature from all the previous. To achieve that we can add an incremental number and a timestamp. Those tow data will be used into the client's logic (and C2S too) to decide if runs command or not.

Each client will run a command if it has a incremental number greater than the last one executed and is passed at most one hour from the timestamp.

### Certificate Revocation
> How can we revoke a certificate to avoid zombies connect to not more secure C2S?

Client initially has an empty certificates-revoked array. Periodically BM sends the signed updated list to all the clients, which use it to synchronize the internal array.
Must pay attenction to don't **cut out yourself** blacklisting all the C2S before order  zombies to move to a new C2S.

## TODO

Find a way to contact all the clients and let they know new ip of a new C2S, without contact them through a previous C2S that could be go down in every moment. (crafted ad-hoc ICMP??)
