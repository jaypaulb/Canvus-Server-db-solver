# Canvus Server DB Solver - Requirements Q&A

## Problem Definition & Context

### 1. Scale & Impact
**Question**: How many Canvus workspaces/canvases are affected by missing assets? What's the typical size of the assets folder?

**Answer**:
```
They have a few thousanc workspaces.  I have no idea how large the asset folder is, assume several terabytes as its media assets and video files / pdfs can be large sometimes.
```

### 2. Asset Types
**Question**: What types of files are typically stored as assets (images, PDFs, videos, documents)? Are there any file size limitations or special handling requirements?

**Answer**:
```
We should check in the documetation for the Canvus Server once we get that from the github as to what the file types are but that assumepion of images, pdfs, videos, documents seem logical.
I don't think there are any special file handling rquirement.
```

### 3. Current Workaround
**Question**: You mentioned re-uploading files makes them reappear - is this the current manual workaround being used? How time-consuming is this process?

**Answer**:
```
Yes this is the manual work around, it is entirely impractical to the point of almost impossible.  There are thousands of canvases with 10's of thousands of potential assets.  We need a scripted/automated process for this action.
```

## Technical Requirements

### 4. Canvus Server Details
**Question**: What version of Canvus Server is KPMG running? Are there any specific API version requirements or authentication limitations?

**Answer**:
```
Canvus server 3.3.0 - the api version in the github repo shared is valid.
```

### 5. Backup Structure
**Question**: When you mention "Father and Grandfather folders" - are these timestamped backup folders (e.g., `backup_2024-01-15`, `backup_2024-01-10`)? What's the typical backup frequency?

**Answer**:
```
Yes - only with more information than backup{date} as the folder name.  I believe they are backing up nightly with 4 gen minimum retention.
```

### 6. File Hashing
**Question**: Do you know what hashing algorithm Canvus uses (MD5, SHA-256, etc.)? This will be important for matching files.

**Answer**:
```
We don't need to know, the api has a hash function built in and in any case we are not going to use it, we are going to query the canvus for the widget list, get the file hash from the widget json, then use that to find the relevent file as the NAME of the file is always {hash_value}.{filetype}
```

### 7. Windows Deployment
**Question**: Will this run as a standalone executable, or do you need an installer? Any specific Windows versions to target?

**Answer**:
```
Stand alone, windows 11 and windows server.  This may need to run as admin in order to copy the backup files to the active assets folder as I think the assets folder is in programdata.
```

## User Experience & Workflow

### 8. User Interface
**Question**: Do you prefer a command-line interface, or would a simple GUI be better for the KPMG team?

**Answer**:
```
I'd prefer a gui but lets get this working from the cli first and we can work on a gui if we have time.  Speed is more important than aesthetics here.
```

### 9. Reporting
**Question**: What level of detail do you need in the report? Just missing files, or also usage statistics, file sizes, etc.?

**Answer**:
```
The report should have the information like
Missing Files
CanvusName - CanvusID
  WidgetName - Hash
  WidgetName - Hash
  WidgetName - Hash
  WidgetName - Hash
CanvusName - CanvusID
  WidgetName - Hash
  WidgetName - Hash
  WidgetName - Hash
  WidgetName - Hash
etc...

This should be accompanied by a matching csv list of just the {hash}.{ext} that can be used for search the backup and restoring the files.
```

### 10. Safety Measures
**Question**: Should the tool have a "dry-run" mode to show what would be copied before actually performing the restoration?

**Answer**:
```
great idea but I don't think its necessary.  The restore non destructive and we have validated that the file is missing in the first stage.
```

## Risk & Constraints

### 11. Server Impact
**Question**: Should the tool avoid making any changes to the Canvus database, or is it acceptable to trigger re-indexing?

**Answer**:
```
My initial concept was to do this via the API and query the API to get the information from the server.  I suspect this is the most "robust" way to do this but the slowest.
The other alternative is to interact with the DB directly but that has the risk of corrupting the DB.  If we did interface with the DB we would need safeguards in place to ensure that this was read only access.
```

### 12. Backup Integrity
**Question**: How do you want to handle cases where a file exists in multiple backup folders with different timestamps?

**Answer**:
```
We use the newest one.
```

### 13. Permissions
**Question**: Will the tool need to handle file permission issues when copying from backup locations?

**Answer**:
```
Not from as the backup location will be mounted as a drive on the machine.  Copy to will need admin rights most probably due to the location in windows of the asset folder.
```

### 14. Timeline
**Question**: What's your target timeline for having this tool ready for use?

**Answer**:
```
As soon as possible.  Deadline is 9th Sept 2025.
```

## Critical Design Questions

### 15. Asset Identification Method
**Question**: Should the tool query the Canvus API to get all asset hashes from the database, then check which files are missing from the filesystem? Or do you want to scan the filesystem first and cross-reference with the database?

**Answer**:
```
Which way will be faster and more robust?  Could they happen in parrellel, scan the assets folder for file names, at the same time as getting the hasses from the api then compare?
```

### 16. Backup Search Strategy
**Question**: When searching multiple backup folders, should it prioritize the most recent backup that contains the file, or do you want to see all available versions?

**Answer**:
```
Most recent - because if the hash is the same it won't matter and we should search new to old.
```

### 17. Error Handling
**Question**: How should the tool handle cases where a file exists in the database but not in any backup location?

**Answer**:
```
igf the hash is in the db but we can't find the file this needs to be note in the final report.
```

### 18. Logging & Audit
**Question**: Do you need detailed logging for compliance/audit purposes, or is basic operation logging sufficient?

**Answer**:
```
option for verbose loging.
```

### 19. Performance Considerations
**Question**: Are there any performance requirements (e.g., must complete within X hours for Y number of assets)?

**Answer**:
```
no.
```

---

## Additional Notes
```
Your_Answer_Here
```
