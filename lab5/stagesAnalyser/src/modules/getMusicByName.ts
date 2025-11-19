export interface ITunesResult {
    resultCount: number;
    results: {
        wrapperType: string;
        artistName: string;
        collectionCensoredName: string;
        trackViewUrl: string;
        artworkUrl100: string;
        collectionId: number;
    }[];
}
