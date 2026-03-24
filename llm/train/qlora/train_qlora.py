#!/usr/bin/env python3
"""QLoRA SFT trainer for commercial JSON-contract dataset."""

from __future__ import annotations

import argparse
import json
from pathlib import Path

import torch
from datasets import load_dataset
from peft import LoraConfig
from transformers import AutoModelForCausalLM, AutoTokenizer, BitsAndBytesConfig
from trl import SFTConfig, SFTTrainer


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description='Train QLoRA adapter on commercial SFT dataset')
    parser.add_argument('--model-name', default='Qwen/Qwen2.5-7B-Instruct')
    parser.add_argument(
        '--train-file',
        default='llm/train/datasets/processed/commercial-sft-json-contract.v1.train.jsonl',
    )
    parser.add_argument(
        '--validation-file',
        default='llm/train/datasets/processed/commercial-sft-json-contract.v1.validation.jsonl',
    )
    parser.add_argument('--output-dir', default='llm/train/qlora/artifacts/qlora-pilot-001')
    parser.add_argument('--max-seq-length', type=int, default=2048)
    parser.add_argument('--learning-rate', type=float, default=2e-4)
    parser.add_argument('--epochs', type=float, default=2.0)
    parser.add_argument('--batch-size', type=int, default=1)
    parser.add_argument('--grad-accum', type=int, default=16)
    parser.add_argument('--warmup-ratio', type=float, default=0.03)
    parser.add_argument('--eval-steps', type=int, default=200)
    parser.add_argument('--save-steps', type=int, default=200)
    parser.add_argument('--logging-steps', type=int, default=20)
    parser.add_argument('--max-train-samples', type=int, default=0)
    parser.add_argument('--max-eval-samples', type=int, default=0)
    parser.add_argument('--seed', type=int, default=42)
    return parser.parse_args()


def build_text_row(row: dict) -> str:
    system = str(row.get('system', '')).strip()
    user = str(row.get('user', '')).strip()
    assistant = row.get('assistant', {})
    assistant_json = json.dumps(assistant, ensure_ascii=False)
    return (
        '<|system|>\n'
        + system
        + '\n<|user|>\n'
        + user
        + '\n<|assistant|>\n'
        + assistant_json
    )


def add_text_column(example: dict) -> dict:
    example['text'] = build_text_row(example)
    return example


def main() -> int:
    args = parse_args()

    out_dir = Path(args.output_dir)
    out_dir.mkdir(parents=True, exist_ok=True)

    dataset = load_dataset(
        'json',
        data_files={
            'train': args.train_file,
            'validation': args.validation_file,
        },
    )

    train_ds = dataset['train']
    eval_ds = dataset['validation']
    if args.max_train_samples > 0:
        train_ds = train_ds.select(range(min(args.max_train_samples, len(train_ds))))
    if args.max_eval_samples > 0:
        eval_ds = eval_ds.select(range(min(args.max_eval_samples, len(eval_ds))))

    train_ds = train_ds.map(add_text_column)
    eval_ds = eval_ds.map(add_text_column)

    bnb_config = BitsAndBytesConfig(
        load_in_4bit=True,
        bnb_4bit_quant_type='nf4',
        bnb_4bit_use_double_quant=True,
        bnb_4bit_compute_dtype=torch.float16,
    )

    tokenizer = AutoTokenizer.from_pretrained(args.model_name, use_fast=True)
    if tokenizer.pad_token is None:
        tokenizer.pad_token = tokenizer.eos_token

    model = AutoModelForCausalLM.from_pretrained(
        args.model_name,
        quantization_config=bnb_config,
        device_map='auto',
        trust_remote_code=True,
    )
    model.config.use_cache = False

    peft_config = LoraConfig(
        r=16,
        lora_alpha=32,
        lora_dropout=0.05,
        bias='none',
        task_type='CAUSAL_LM',
        target_modules=['q_proj', 'k_proj', 'v_proj', 'o_proj', 'gate_proj', 'up_proj', 'down_proj'],
    )

    training_args = SFTConfig(
        output_dir=str(out_dir),
        num_train_epochs=args.epochs,
        per_device_train_batch_size=args.batch_size,
        per_device_eval_batch_size=1,
        gradient_accumulation_steps=args.grad_accum,
        learning_rate=args.learning_rate,
        warmup_ratio=args.warmup_ratio,
        max_seq_length=args.max_seq_length,
        logging_steps=args.logging_steps,
        save_steps=args.save_steps,
        eval_steps=args.eval_steps,
        eval_strategy='steps',
        save_strategy='steps',
        save_total_limit=2,
        report_to='none',
        bf16=False,
        fp16=True,
        dataloader_num_workers=2,
        gradient_checkpointing=True,
        seed=args.seed,
    )

    trainer = SFTTrainer(
        model=model,
        args=training_args,
        train_dataset=train_ds,
        eval_dataset=eval_ds,
        peft_config=peft_config,
        processing_class=tokenizer,
        dataset_text_field='text',
    )

    trainer.train()
    trainer.save_model(str(out_dir))
    tokenizer.save_pretrained(str(out_dir))
    print(f'Saved adapter to {out_dir}')
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
