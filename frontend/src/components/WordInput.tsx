import {useState} from "react";

interface Props {
    disabled: boolean;
    onSubmit: (word: string) => void;
}


export function WordInput({ disabled, onSubmit}: Props) {
    const [value, setValue] = useState("");

    function handleSubmit(e: React.FormEvent) {
        //prevents the page from refreshing on form submission
        e.preventDefault();
        const word = value.trim().toLowerCase();
        if( !word ) return;

        onSubmit(word);
        setValue("");
    }

    return (
        <form onSubmit={handleSubmit} style = {{marginTop: 16}}>
            <input 
                type="text"
                value={value}
                disabled = {disabled}
                onChange={(e) => setValue(e.target.value)}
                placeholder="Enter Word..."
                style={{padding: 8, width: "70%"}}
            />
            <button 
                type="submit"
                disabled={disabled}
                style={{marginLeft: 8, padding: 8}}
            >Submit</button>
        </form>
    );
}